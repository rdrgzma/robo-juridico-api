package bridge

import (
	"strings"
	"testing"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/calculator"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/licensing"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/prazos"
	"github.com/shopspring/decimal"
)

func TestCoreBridgeCompleto(t *testing.T) {
	// 1. Gera par de chaves ECDSA para simular licenciamento
	privKey, _, pubPEM, err := licensing.GerarChavesECDSA()
	if err != nil {
		t.Fatalf("Erro ao gerar chaves ECDSA: %v", err)
	}

	cfg := ConfigDesktop{
		Versao:          "1.0.0-test",
		ChavePublicaPEM: pubPEM,
	}

	core, err := NovoCoreBridge(cfg)
	if err != nil {
		t.Fatalf("Erro ao inicializar CoreBridge: %v", err)
	}

	// 2. Testa ObterInfoSistema
	info := core.ObterInfoSistema()
	if info.HardwareID == "" {
		t.Errorf("Esperava HardwareID preenchido")
	}
	if info.TotalIndices == 0 {
		t.Errorf("Esperava indices economicos carregados")
	}

	// 3. Testa Validação de CNJ Módulo 97
	cnjValido := "5010754-07.2024.8.13.0625"
	p := core.ValidarCNJ(cnjValido)
	if p == nil || !p.Valido {
		t.Fatalf("Esperava processo valido para CNJ com digito correto")
	}

	// 4. Testa Licenciamento: Deve bloquear cálculo sem licença ativa
	reqCalc := calculator.CalculoJudicialRequest{
		DataAtualizacao: "2024-01-01",
		IndexadorCodigo: "IPCA",
		JurosMora: calculator.JurosMoraConfig{
			Tipo:             calculator.JurosConvencionalMensal,
			PercentualMensal: decimal.NewFromFloat(1.0),
			Modo:             calculator.CapitalizacaoSimples,
		},
		Parcelas: []calculator.ParcelaItem{
			{
				NumeroItem:     1,
				Descricao:      "Dano Material",
				ValorOriginal:  decimal.NewFromFloat(1000.00),
				DataVencimento: "2023-01-01",
				Tipo:           calculator.ParcelaDebito,
			},
		},
	}

	_, errBloq := core.CalcularLiquidacao(reqCalc)
	if errBloq != ErrLicencaObrigatoria {
		t.Fatalf("Esperava ErrLicencaObrigatoria, obteve: %v", errBloq)
	}

	// 5. Ativação de Licença Válida com todos os recursos
	emissor := licensing.NovoEmissor(privKey)
	tokenLic, err := emissor.Emitir(licensing.DadosLicenca{
		IDCliente:   "cli_advocacia_sp",
		NomeTitular: "Dra. Maria Advocacia",
		Plano:       "ilimitado",
		HardwareID:  info.HardwareID,
		ValidaAte:   time.Now().Add(365 * 24 * time.Hour),
		Recursos: licensing.RecursosLicenca{
			CalculoJudicialCompleto: true,
			RevisaoBancaria:         true,
			ContagemPrazos:          true,
			ConsultaProcessual:      true,
			ExportarExcel:           true,
			ExportarLaudoPDF:        true,
			AtualizacaoAutomatica:   true,
		},
	})
	if err != nil {
		t.Fatalf("Erro ao emitir token de teste: %v", err)
	}

	_, errAtiv := core.AtivarLicenca(tokenLic)
	if errAtiv != nil {
		t.Fatalf("Erro ao ativar licenca no CoreBridge: %v", errAtiv)
	}

	// 6. Agora o cálculo deve passar com sucesso
	respCalc, errCalc := core.CalcularLiquidacao(reqCalc)
	if errCalc != nil {
		t.Fatalf("Erro ao executar calculo apos ativar licenca: %v", errCalc)
	}
	if respCalc.Resumo.TotalGeralDevido.LessThanOrEqual(decimal.Zero) {
		t.Errorf("Total geral devido deve ser maior que zero")
	}

	// 7. Testa Exportação CSV
	csv, errCSV := core.ExportarCalculoCSV(respCalc)
	if errCSV != nil {
		t.Fatalf("Erro ao exportar CSV: %v", errCSV)
	}
	if !strings.Contains(csv, "Dano Material") {
		t.Errorf("CSV nao contem o item de Dano Material")
	}

	// 8. Testa Revisão Bancária (Financiamento Tabela Price)
	reqRev := calculator.CalculoFinanciamentoRequest{
		ValorFinanciado:      decimal.NewFromFloat(50000.00),
		NumeroParcelas:       48,
		TaxaContratadaMensal: decimal.NewFromFloat(2.5),
		TaxaMercadoMensal:    decimal.NewFromFloat(1.5),
		SistemaAmortizacao:   calculator.SistemaPrice,
	}
	respRev, errRev := core.CalcularRevisional(reqRev)
	if errRev != nil {
		t.Fatalf("Erro ao calcular revisional: %v", errRev)
	}
	if respRev.DiferencaTotalSimples.LessThanOrEqual(decimal.Zero) {
		t.Errorf("Esperava diferenca a restituir positiva na revisional")
	}

	// 9. Testa Contagem de Prazo Processual (CPC)
	reqPrazo := prazos.CalculoPrazoRequest{
		DataDisponibilizacao: "2024-05-02", // Quinta-feira
		DiasPrazo:            15,
		Rito:                 "CPC_UTEIS",
		TipoIntimacao:        "DIARIO_ELETRONICO",
	}
	respPrazo, errPrazo := core.CalcularPrazo(reqPrazo)
	if errPrazo != nil {
		t.Fatalf("Erro ao calcular prazo processual: %v", errPrazo)
	}
	if !respPrazo.Sucesso {
		t.Errorf("Esperava prazo com sucesso")
	}
	if respPrazo.DataVencimentoFatal == "" {
		t.Errorf("Esperava data de vencimento fatal preenchida")
	}
}
