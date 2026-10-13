package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/cache"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/cnj"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/domain"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/fontes/comunica"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/fontes/datajud"
)

const (
	ValidadeDataJud = 6 * time.Hour
	ValidadeDiario  = 30 * time.Minute
)

type Orquestrador struct {
	datajudClient  *datajud.ClienteDataJud
	comunicaClient *comunica.ClienteComunica
	cache          *cache.CacheTTL
}

func NovoOrquestrador(dj *datajud.ClienteDataJud, com *comunica.ClienteComunica, c *cache.CacheTTL) *Orquestrador {
	return &Orquestrador{
		datajudClient:  dj,
		comunicaClient: com,
		cache:          c,
	}
}

func (o *Orquestrador) ConsultarProcessoDetalhado(ctx context.Context, numero string, comPublicacoes bool, modo string) (*domain.Processo, []string, error) {
	procCNJ := cnj.Parse(numero)
	if !procCNJ.Valido {
		return nil, nil, domain.ErrNumeroInvalido
	}

	var avisos []string

	// Modo rápido: somente DJEN
	if modo == "rapido" {
		pubs, err := o.buscarDJENComCache(ctx, procCNJ.Digitos)
		if err != nil {
			return nil, nil, err
		}
		if len(pubs) == 0 {
			return nil, nil, nil
		}
		proc := o.montarProcessoDasPublicacoes(procCNJ, pubs)
		return proc, avisos, nil
	}

	// Modo completo: DataJud e DJEN em paralelo
	var wg sync.WaitGroup
	var procDataJud *domain.Processo
	var errDataJud error
	var pubsDJEN []domain.Publicacao
	var errDJEN error

	wg.Add(1)
	go func() {
		defer wg.Done()
		procDataJud, errDataJud = o.buscarDataJudComCache(ctx, procCNJ.Formatado, procCNJ.Digitos)
	}()

	if comPublicacoes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pubsDJEN, errDJEN = o.buscarDJENComCache(ctx, procCNJ.Digitos)
		}()
	}

	wg.Wait()

	if errDataJud != nil {
		avisos = append(avisos, "O DataJud não respondeu a tempo, então o histórico profundo de andamentos ficou de fora.")
	}
	if errDJEN != nil && comPublicacoes {
		avisos = append(avisos, "O diário eletrônico (DJEN) não respondeu, então as publicações ficaram de fora.")
	}

	// Nenhuma fonte encontrou nada
	if procDataJud == nil && len(pubsDJEN) == 0 {
		return nil, avisos, nil
	}

	// Apenas o DJEN encontrou (ex: processo recém-distribuído)
	if procDataJud == nil && len(pubsDJEN) > 0 {
		proc := o.montarProcessoDasPublicacoes(procCNJ, pubsDJEN)
		return proc, avisos, nil
	}

	// DataJud encontrou: mescla com as publicações do diário se houver
	if len(pubsDJEN) > 0 {
		o.juntarPublicacoes(procDataJud, pubsDJEN)
	}

	return procDataJud, avisos, nil
}

func (o *Orquestrador) buscarDataJudComCache(ctx context.Context, formatado, digitos string) (*domain.Processo, error) {
	chave := fmt.Sprintf("datajud:%s", digitos)
	val, err := o.cache.ObterOuBuscar(chave, ValidadeDataJud, func() (any, error) {
		return o.datajudClient.ConsultarProcesso(ctx, formatado)
	})
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}
	return val.(*domain.Processo), nil
}

func (o *Orquestrador) buscarDJENComCache(ctx context.Context, digitos string) ([]domain.Publicacao, error) {
	chave := fmt.Sprintf("diario:%s", digitos)
	val, err := o.cache.ObterOuBuscar(chave, ValidadeDiario, func() (any, error) {
		return o.comunicaClient.PublicacoesDoProcesso(ctx, digitos, time.Time{}, time.Time{})
	})
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}
	return val.([]domain.Publicacao), nil
}

func (o *Orquestrador) montarProcessoDasPublicacoes(procCNJ *cnj.ProcessoCNJ, pubs []domain.Publicacao) *domain.Processo {
	sort.Slice(pubs, func(i, j int) bool {
		return pubs[i].DataDisponibilizacao > pubs[j].DataDisponibilizacao
	})
	primeira := pubs[0]

	partesMap := make(map[string]domain.Parte)
	advsMap := make(map[string]domain.Advogado)

	var movs []domain.Movimentacao
	for _, pub := range pubs {
		for _, parte := range pub.Partes {
			if _, ok := partesMap[parte.Nome]; !ok || partesMap[parte.Nome].Polo == domain.PoloTerceiro {
				partesMap[parte.Nome] = parte
			}
		}
		for _, adv := range pub.Advogados {
			chaveAdv := fmt.Sprintf("%s|%s", adv.Nome, adv.OAB)
			advsMap[chaveAdv] = adv
		}

		t, df := comunica.ParseDataDJEN(pub.DataDisponibilizacao)
		movs = append(movs, domain.Movimentacao{
			Data:          t,
			DataFormatada: df,
			Tipo:          string(comunica.ClassificarTipo(pub.TipoComunicacao)),
			Descricao:     pub.TipoComunicacao,
			Complemento:   pub.Texto,
			Orgao:         pub.Orgao,
			Fonte:         string(domain.FonteComunica),
			Link:          pub.Link,
			IDExterno:     fmt.Sprintf("comunica-%s", pub.ID),
		})
	}

	var partes []domain.Parte
	for _, p := range partesMap {
		partes = append(partes, p)
	}
	var advs []domain.Advogado
	for _, a := range advsMap {
		advs = append(advs, a)
	}

	proc := &domain.Processo{
		Numero:                         procCNJ.Formatado,
		NumeroDigitos:                  procCNJ.Digitos,
		Tribunal:                       procCNJ.Tribunal,
		Segmento:                       procCNJ.NomeSegmento,
		Classe:                         primeira.Classe,
		OrgaoJulgador:                  primeira.Orgao,
		Link:                           primeira.Link,
		DataUltimaMovimentacao:         primeira.DataDisponibilizacao,
		DataUltimaMovimentacaoFormatada: primeira.DataFormatada,
		Partes:                         partes,
		Advogados:                      advs,
		Movimentacoes:                  movs,
		Publicacoes:                    pubs,
		Fontes:                         []string{string(domain.FonteComunica)},
		Polos: domain.Polos{
			Ativo:     filtrarPolo(partes, domain.PoloAtivo),
			Passivo:   filtrarPolo(partes, domain.PoloPassivo),
			Terceiros: filtrarPolo(partes, domain.PoloTerceiro),
		},
		Totais: domain.Totais{
			Partes:        len(partes),
			Advogados:     len(advs),
			Movimentacoes: len(movs),
			Publicacoes:   len(pubs),
		},
	}
	proc.Titulo = montarTitulo(proc.Polos)
	return proc
}

func (o *Orquestrador) juntarPublicacoes(proc *domain.Processo, pubs []domain.Publicacao) {
	proc.Publicacoes = pubs
	temFonteComunica := false
	for _, f := range proc.Fontes {
		if f == string(domain.FonteComunica) {
			temFonteComunica = true
			break
		}
	}
	if !temFonteComunica {
		proc.Fontes = append(proc.Fontes, string(domain.FonteComunica))
	}

	if proc.Link == "" && len(pubs) > 0 {
		proc.Link = pubs[0].Link
	}

	// Deduplicação de movimentos
	jaTem := make(map[string]bool)
	for _, m := range proc.Movimentacoes {
		if m.IDExterno != "" {
			jaTem[m.IDExterno] = true
		}
	}

	for _, pub := range pubs {
		idExt := fmt.Sprintf("comunica-%s", pub.ID)
		if !jaTem[idExt] {
			t, df := comunica.ParseDataDJEN(pub.DataDisponibilizacao)
			proc.Movimentacoes = append(proc.Movimentacoes, domain.Movimentacao{
				Data:          t,
				DataFormatada: df,
				Tipo:          string(comunica.ClassificarTipo(pub.TipoComunicacao)),
				Descricao:     pub.TipoComunicacao,
				Complemento:   pub.Texto,
				Orgao:         pub.Orgao,
				Fonte:         string(domain.FonteComunica),
				Link:          pub.Link,
				IDExterno:     idExt,
			})
			jaTem[idExt] = true
		}
	}

	// Ordena movimentos do mais recente para o mais antigo
	sort.Slice(proc.Movimentacoes, func(i, j int) bool {
		return proc.Movimentacoes[i].Data.After(proc.Movimentacoes[j].Data)
	})

	// Enriquece partes e advogados do DataJud com o que veio do diário
	nomesPartes := make(map[string]bool)
	for _, p := range proc.Partes {
		nomesPartes[p.Nome] = true
	}
	for _, pub := range pubs {
		for _, parte := range pub.Partes {
			if !nomesPartes[parte.Nome] {
				nomesPartes[parte.Nome] = true
				proc.Partes = append(proc.Partes, parte)
			}
		}
	}

	advsConhecidos := make(map[string]bool)
	for _, a := range proc.Advogados {
		advsConhecidos[fmt.Sprintf("%s|%s", a.Nome, a.OAB)] = true
	}
	for _, pub := range pubs {
		for _, adv := range pub.Advogados {
			chaveAdv := fmt.Sprintf("%s|%s", adv.Nome, adv.OAB)
			if !advsConhecidos[chaveAdv] {
				advsConhecidos[chaveAdv] = true
				proc.Advogados = append(proc.Advogados, adv)
			}
		}
	}

	// Atualiza totais e polos
	proc.Polos = domain.Polos{
		Ativo:     filtrarPolo(proc.Partes, domain.PoloAtivo),
		Passivo:   filtrarPolo(proc.Partes, domain.PoloPassivo),
		Terceiros: filtrarPolo(proc.Partes, domain.PoloTerceiro),
	}
	proc.Totais = domain.Totais{
		Partes:        len(proc.Partes),
		Advogados:     len(proc.Advogados),
		Movimentacoes: len(proc.Movimentacoes),
		Publicacoes:   len(proc.Publicacoes),
	}
	proc.Titulo = montarTitulo(proc.Polos)
}

func (o *Orquestrador) Existe(ctx context.Context, numero string) (bool, string, error) {
	procCNJ := cnj.Parse(numero)
	if !procCNJ.Valido {
		return false, "", domain.ErrNumeroInvalido
	}

	// Checa DataJud primeiro com timeout curto
	ctxDJ, cancel := context.WithTimeout(ctx, 10*time.Second)
	p, err := o.datajudClient.ConsultarProcesso(ctxDJ, procCNJ.Formatado)
	cancel()
	if err == nil && p != nil {
		return true, string(domain.FonteDataJud), nil
	}

	// Checa DJEN
	ctxCom, cancelCom := context.WithTimeout(ctx, 10*time.Second)
	pubs, err := o.comunicaClient.PublicacoesDoProcesso(ctxCom, procCNJ.Digitos, time.Time{}, time.Time{})
	cancelCom()
	if err == nil && len(pubs) > 0 {
		return true, string(domain.FonteComunica), nil
	}

	return false, "", nil
}

func (o *Orquestrador) BuscarPorOAB(ctx context.Context, oabNumero, oabUF, tribunal string, dias int) ([]domain.ResumoProcesso, error) {
	fim := time.Now().UTC()
	inicio := fim.AddDate(0, 0, -dias)

	pubs, err := o.comunicaClient.PublicacoesPorOAB(ctx, oabNumero, oabUF, inicio, fim, 1)
	if err != nil {
		return nil, err
	}

	alvoTribunal := strings.ToUpper(strings.TrimSpace(tribunal))
	agrupados := make(map[string][]domain.Publicacao)

	for _, pub := range pubs {
		pCNJ := cnj.Parse(pub.NumeroProcesso)
		if !pCNJ.Valido {
			continue
		}
		if alvoTribunal != "" && pCNJ.Tribunal != alvoTribunal {
			continue
		}
		agrupados[pCNJ.Formatado] = append(agrupados[pCNJ.Formatado], pub)
	}

	var resumos []domain.ResumoProcesso
	for num, lista := range agrupados {
		sort.Slice(lista, func(i, j int) bool {
			return lista[i].DataDisponibilizacao > lista[j].DataDisponibilizacao
		})
		primeira := lista[0]
		pCNJ := cnj.Parse(num)

		resumos = append(resumos, domain.ResumoProcesso{
			Numero:                         num,
			Titulo:                         primeira.Classe,
			Tribunal:                       pCNJ.Tribunal,
			Classe:                         primeira.Classe,
			OrgaoJulgador:                  primeira.Orgao,
			DataUltimaMovimentacao:         primeira.DataDisponibilizacao,
			DataUltimaMovimentacaoFormatada: primeira.DataFormatada,
			Publicacoes:                    len(lista),
		})
	}

	sort.Slice(resumos, func(i, j int) bool {
		return resumos[i].DataUltimaMovimentacao > resumos[j].DataUltimaMovimentacao
	})

	return resumos, nil
}

func (o *Orquestrador) PublicacoesPorOAB(ctx context.Context, oabNumero, oabUF string, inicio, fim time.Time, paginas int) ([]domain.Publicacao, error) {
	return o.comunicaClient.PublicacoesPorOAB(ctx, oabNumero, oabUF, inicio, fim, paginas)
}

func filtrarPolo(partes []domain.Parte, polo domain.Polo) []domain.Parte {
	var filtradas []domain.Parte
	for _, p := range partes {
		if p.Polo == polo {
			filtradas = append(filtradas, p)
		}
	}
	return filtradas
}

func montarTitulo(polos domain.Polos) string {
	var ativos []string
	for _, p := range polos.Ativo {
		ativos = append(ativos, p.Nome)
	}
	var passivos []string
	for _, p := range polos.Passivo {
		passivos = append(passivos, p.Nome)
	}

	ativoStr := "—"
	if len(ativos) > 0 {
		ativoStr = strings.Join(ativos, ", ")
	}
	passivoStr := "—"
	if len(passivos) > 0 {
		passivoStr = strings.Join(passivos, ", ")
	}

	return fmt.Sprintf("%s × %s", ativoStr, passivoStr)
}
