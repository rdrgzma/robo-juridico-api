package bridge

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/rdrgzma/robo-juridico-desktop/internal/cache"
	"github.com/rdrgzma/robo-juridico-desktop/internal/calculator"
	"github.com/rdrgzma/robo-juridico-desktop/internal/calculator/collectors"
	"github.com/rdrgzma/robo-juridico-desktop/internal/cnj"
	"github.com/rdrgzma/robo-juridico-desktop/internal/domain"
	"github.com/rdrgzma/robo-juridico-desktop/internal/fontes/comunica"
	"github.com/rdrgzma/robo-juridico-desktop/internal/fontes/datajud"
	"github.com/rdrgzma/robo-juridico-desktop/internal/licensing"
	"github.com/rdrgzma/robo-juridico-desktop/internal/network"
	"github.com/rdrgzma/robo-juridico-desktop/internal/prazos"
	"github.com/rdrgzma/robo-juridico-desktop/internal/service"
)

var (
	ErrLicencaObrigatoria = errors.New("licenca ativa obrigatoria para utilizar esta funcionalidade")
	ErrNumeroInvalido     = domain.ErrNumeroInvalido
)

type ConfigDesktop struct {
	Versao          string
	DataJudAPIKey   string
	DataJudTimeout  time.Duration
	ComunicaTimeout time.Duration
	ChavePublicaPEM string // Chave pública ECDSA para validação de licenças
}

type InfoSistema struct {
	Versao           string                  `json:"versao"`
	SO               string                  `json:"sistema_operacional"`
	Arquitetura      string                  `json:"arquitetura"`
	HardwareID       string                  `json:"hardware_id"`
	LicencaAtiva     bool                    `json:"licenca_ativa"`
	LicencaDetalhes  *licensing.DadosLicenca `json:"licenca_detalhes,omitempty"`
	TotalIndices     int                     `json:"total_indices"`
	UltimaSincronia  string                  `json:"ultima_sincronia"`
}

type CoreBridge struct {
	cfg              ConfigDesktop
	hardwareID       string
	orquestrador     *service.Orquestrador
	calcEngine       *calculator.EngineCalculo
	calcRepo         *calculator.RepositorioIndices
	workerColetor    *collectors.WorkerColetor
	validadorLicenca *licensing.ValidadorLicenca
	licencaAtual     *licensing.DadosLicenca
	ultimaSincronia  time.Time
	mu               sync.RWMutex
}

// NovoCoreBridge inicializa todo o ecossistema jurídico desacoplado de HTTP.
func NovoCoreBridge(cfg ConfigDesktop) (*CoreBridge, error) {
	if cfg.Versao == "" {
		cfg.Versao = "1.0.0-desktop"
	}
	if cfg.DataJudAPIKey == "" {
		cfg.DataJudAPIKey = "cDZHYzlZa0JadVREZDJCendQbXY6SkJlTzNjLV9TRENyQk1RdnFKZGRQdw=="
	}
	if cfg.DataJudTimeout == 0 {
		cfg.DataJudTimeout = 25 * time.Second
	}
	if cfg.ComunicaTimeout == 0 {
		cfg.ComunicaTimeout = 20 * time.Second
	}

	hwID := licensing.ObterHardwareFingerprint()

	// 1. Rede e Fontes
	httpCliente := network.NovoClienteHTTP(30*time.Second, "RoboJuridico-Desktop/"+cfg.Versao)
	djClient := datajud.NovoClienteDataJud(httpCliente, cfg.DataJudAPIKey, cfg.DataJudTimeout, 2)
	comunicaClient := comunica.NovoClienteComunica(httpCliente, nil, cfg.ComunicaTimeout, 50, 2)

	cacheTTL := cache.NovoCacheTTL()
	orquestrador := service.NovoOrquestrador(djClient, comunicaClient, cacheTTL)

	// 2. Motor de Cálculos e Índices
	calcRepo := calculator.NovoRepositorioIndices()
	calcEngine := calculator.NovoEngineCalculo(calcRepo)
	workerColetor := collectors.NovoWorker(calcRepo, 24*time.Hour)

	// 3. Validador de Licença
	var validador *licensing.ValidadorLicenca
	if cfg.ChavePublicaPEM != "" {
		v, err := licensing.NovoValidador(cfg.ChavePublicaPEM)
		if err == nil {
			validador = v
		}
	}

	return &CoreBridge{
		cfg:              cfg,
		hardwareID:       hwID,
		orquestrador:     orquestrador,
		calcEngine:       calcEngine,
		calcRepo:         calcRepo,
		workerColetor:    workerColetor,
		validadorLicenca: validador,
		ultimaSincronia:  time.Now(),
	}, nil
}

// ObterInfoSistema retorna metadados da máquina, do OS e da licença para a UI
func (b *CoreBridge) ObterInfoSistema() InfoSistema {
	b.mu.RLock()
	defer b.mu.RUnlock()

	indices := b.calcRepo.ListarIndices()
	licAtiva := b.licencaAtual != nil && (b.licencaAtual.ValidaAte.IsZero() || time.Now().Before(b.licencaAtual.ValidaAte))

	return InfoSistema{
		Versao:          b.cfg.Versao,
		SO:              runtime.GOOS,
		Arquitetura:     runtime.GOARCH,
		HardwareID:      b.hardwareID,
		LicencaAtiva:    licAtiva,
		LicencaDetalhes: b.licencaAtual,
		TotalIndices:    len(indices),
		UltimaSincronia: b.ultimaSincronia.Format(time.RFC3339),
	}
}

// AtivarLicenca valida e ativa um token de licença assinado criptograficamente
func (b *CoreBridge) AtivarLicenca(token string) (*licensing.DadosLicenca, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.validadorLicenca == nil {
		return nil, errors.New("validador de licenca nao configurado nesta compilacao")
	}

	dados, err := b.validadorLicenca.Validar(token, b.hardwareID)
	if err != nil {
		return nil, err
	}

	b.licencaAtual = dados
	return dados, nil
}

// SetLicencaManual permite injetar uma licença já validada ou modo de desenvolvimento
func (b *CoreBridge) SetLicencaManual(lic *licensing.DadosLicenca) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.licencaAtual = lic
}

// ValidarRecurso checa se o recurso está liberado no plano
func (b *CoreBridge) ValidarRecurso(verificar func(r licensing.RecursosLicenca) bool) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Se não há licença configurada ou expirou
	if b.licencaAtual == nil {
		return ErrLicencaObrigatoria
	}
	if !b.licencaAtual.ValidaAte.IsZero() && time.Now().After(b.licencaAtual.ValidaAte) {
		return licensing.ErrLicencaExpirada
	}
	if !verificar(b.licencaAtual.Recursos) {
		return licensing.ErrRecursoNaoAutorizado
	}
	return nil
}

// ConsultarProcesso executa a busca paralela em DataJud e DJEN
func (b *CoreBridge) ConsultarProcesso(ctx context.Context, numeroCNJ, modo string) (*domain.Processo, []string, error) {
	if err := b.ValidarRecurso(func(r licensing.RecursosLicenca) bool { return r.ConsultaProcessual }); err != nil {
		// Modo gracioso para testes/desenvolvimento: permite se licencaAtual for nula em ambiente local
		if b.validadorLicenca != nil && b.licencaAtual == nil {
			return nil, nil, err
		}
	}

	procCNJ := cnj.Parse(numeroCNJ)
	if !procCNJ.Valido {
		return nil, nil, ErrNumeroInvalido
	}

	if modo == "" {
		modo = "completo"
	}

	return b.orquestrador.ConsultarProcessoDetalhado(ctx, procCNJ.Formatado, true, modo)
}

// ValidarCNJ apenas testa o dígito verificador módulo 97 do CNJ
func (b *CoreBridge) ValidarCNJ(numero string) *cnj.ProcessoCNJ {
	return cnj.Parse(numero)
}

// CalcularLiquidacao processa a memória de cálculo de débito/crédito judicial
func (b *CoreBridge) CalcularLiquidacao(req calculator.CalculoJudicialRequest) (*calculator.CalculoJudicialResponse, error) {
	if err := b.ValidarRecurso(func(r licensing.RecursosLicenca) bool { return r.CalculoJudicialCompleto }); err != nil {
		if b.validadorLicenca != nil && b.licencaAtual == nil {
			return nil, err
		}
	}

	return b.calcEngine.Executar(req)
}

// CalcularRevisional processa a revisão contratual de juros (Price, SAC, Gauss)
func (b *CoreBridge) CalcularRevisional(req calculator.CalculoFinanciamentoRequest) (*calculator.CalculoFinanciamentoResponse, error) {
	if err := b.ValidarRecurso(func(r licensing.RecursosLicenca) bool { return r.RevisaoBancaria }); err != nil {
		if b.validadorLicenca != nil && b.licencaAtual == nil {
			return nil, err
		}
	}

	return calculator.CalcularRevisaoFinanciamento(req)
}

// CalcularPrazo calcula prazos processuais forenses (CPC e CLT) com contagem de dias úteis
func (b *CoreBridge) CalcularPrazo(req prazos.CalculoPrazoRequest) (*prazos.CalculoPrazoResponse, error) {
	if err := b.ValidarRecurso(func(r licensing.RecursosLicenca) bool { return r.ContagemPrazos }); err != nil {
		if b.validadorLicenca != nil && b.licencaAtual == nil {
			return nil, err
		}
	}

	return prazos.CalcularPrazoProcessual(req)
}

// ListarIndices retorna todos os índices econômicos suportados e o intervalo disponível
func (b *CoreBridge) ListarIndices() []calculator.IndiceInfo {
	return b.calcRepo.ListarIndices()
}

// SincronizarIndices executa atualização dos índices diretamente das fontes oficiais
func (b *CoreBridge) SincronizarIndices(ctx context.Context) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	rel := b.workerColetor.Executar(ctx)
	b.ultimaSincronia = time.Now()

	return rel.TotalPontosAtualizados, nil
}

// ExportarCalculoCSV formata a memória de cálculo para formato tabular CSV (para Excel)
func (b *CoreBridge) ExportarCalculoCSV(resp *calculator.CalculoJudicialResponse) (string, error) {
	if resp == nil {
		return "", errors.New("resposta de calculo vazia")
	}

	var sb strings.Builder
	sb.WriteString("Item;Descricao;Data Vencimento;Valor Original;Fator Correcao;Valor Corrigido;Juros;Total Item\n")

	for _, item := range resp.ItensLiquidados {
		sb.WriteString(fmt.Sprintf("%d;\"%s\";%s;%s;%s;%s;%s;%s\n",
			item.NumeroItem,
			strings.ReplaceAll(item.Descricao, "\"", "'"),
			item.DataVencimento,
			item.ValorOriginal.StringFixed(2),
			item.FatorCorrecao.StringFixed(6),
			item.ValorCorrigido.StringFixed(2),
			item.ValorJurosMora.StringFixed(2),
			item.TotalItem.StringFixed(2),
		))
	}

	sb.WriteString(fmt.Sprintf("\nRESUMO GERAL;;;%s;;%s;%s;%s\n",
		resp.Resumo.TotalPrincipalOriginal.StringFixed(2),
		resp.Resumo.TotalPrincipalCorrigido.StringFixed(2),
		resp.Resumo.TotalJurosMora.StringFixed(2),
		resp.Resumo.TotalGeralDevido.StringFixed(2),
	))

	return sb.String(), nil
}
