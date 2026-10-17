package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
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

	avisos := []string{}

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

	// Modo completo: DataJud e DJEN em paralelo com fast-fallback se o DataJud demorar demais
	type resDJ struct {
		proc *domain.Processo
		err  error
	}
	chDJ := make(chan resDJ, 1)

	go func() {
		// Usa contexto desacoplado do cancelamento imediato para aquecer o cache em background caso o fast-fallback retorne antes
		ctxBg, cancelBg := context.WithTimeout(context.WithoutCancel(ctx), 25*time.Second)
		defer cancelBg()
		p, err := o.buscarDataJudComCache(ctxBg, procCNJ.Formatado, procCNJ.Digitos)
		chDJ <- resDJ{proc: p, err: err}
	}()

	var pubsDJEN []domain.Publicacao
	var errDJEN error
	if comPublicacoes {
		pubsDJEN, errDJEN = o.buscarDJENComCache(ctx, procCNJ.Digitos)
	}

	var procDataJud *domain.Processo
	var errDataJud error

	// Se o DJEN já trouxe publicações rapidamente, aguarda o DataJud por no máximo mais 4.5s (evita travar 27s em tribunais lentos)
	timeoutEsperaDJ := 12 * time.Second
	if len(pubsDJEN) > 0 {
		timeoutEsperaDJ = 4500 * time.Millisecond
	}

	select {
	case r := <-chDJ:
		procDataJud = r.proc
		errDataJud = r.err
	case <-time.After(timeoutEsperaDJ):
		avisos = append(avisos, "O DataJud demorou além do limite rápido; dados consolidados via DJEN (o histórico completo está sendo sincronizado em cache).")
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}

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

	// Apenas o DJEN encontrou (ou respondeu dentro do SLA rápido)
	if procDataJud == nil && len(pubsDJEN) > 0 {
		proc := o.montarProcessoDasPublicacoes(procCNJ, pubsDJEN)
		return proc, avisos, nil
	}

	// DataJud encontrou: clona superficialmente antes de mesclar para não poluir ponteiro em cache
	procFinal := *procDataJud
	if len(pubsDJEN) > 0 {
		o.juntarPublicacoes(&procFinal, pubsDJEN)
	}

	return &procFinal, avisos, nil
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
	comarca := ""
	assunto := ""
	var valorCausa *float64

	var movs []domain.Movimentacao
	for _, pub := range pubs {
		if comarca == "" {
			comarca = comunica.InferirComarcaPorOrigem(procCNJ.Tribunal, procCNJ.OrigemCodigo, pub.Orgao, pub.Texto)
		}
		if valorCausa == nil {
			valorCausa = comunica.ExtrairValorCausaDoTexto(pub.Texto)
		}
		if assunto == "" && pub.TipoDocumento != "" && !strings.EqualFold(pub.TipoDocumento, pub.TipoComunicacao) {
			assunto = pub.TipoDocumento
		}
		for _, parte := range pub.Partes {
			nomeLimpo := comunica.LimparNomeParte(parte.Nome)
			if !comunica.EhNomeParteValido(nomeLimpo) {
				continue
			}
			parte.Nome = nomeLimpo
			chave := strings.ToUpper(nomeLimpo)
			if _, ok := partesMap[chave]; !ok || partesMap[chave].Polo == domain.PoloTerceiro {
				partesMap[chave] = parte
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
			LinkPeca:      pub.LinkPeca,
			IDExterno:     fmt.Sprintf("comunica-%s", pub.ID),
		})
	}

	if comarca == "" {
		comarca = comunica.InferirComarcaPorOrigem(procCNJ.Tribunal, procCNJ.OrigemCodigo, primeira.Orgao, "")
	}
	if assunto == "" {
		assunto = primeira.Classe
	}

	partes := []domain.Parte{}
	for _, p := range partesMap {
		partes = append(partes, p)
	}
	advs := []domain.Advogado{}
	for _, a := range advsMap {
		advs = append(advs, a)
	}

	var dataAjuizamento *time.Time
	dataAjuizamentoFormatada := ""
	if len(movs) > 0 {
		maisAntiga := movs[len(movs)-1].Data
		if !maisAntiga.IsZero() {
			dataAjuizamento = &maisAntiga
			dataAjuizamentoFormatada = maisAntiga.Format("02/01/2006")
		}
	}

	grauInferido := comunica.InferirGrau(primeira.Orgao, primeira.Classe)
	linkTribunal := cnj.GerarLinkTribunal(procCNJ.Formatado, procCNJ.Tribunal, grauInferido, procCNJ.OrigemCodigo)
	if strings.Contains(primeira.Link, "proc_resultado.jsp") {
		linkTribunal = primeira.Link
	}

	linkPecaMaisRecente := ""
	for _, pub := range pubs {
		if pub.LinkPeca != "" {
			linkPecaMaisRecente = pub.LinkPeca
			break
		}
	}

	proc := &domain.Processo{
		Numero:                         procCNJ.Formatado,
		NumeroDigitos:                  procCNJ.Digitos,
		Tribunal:                       procCNJ.Tribunal,
		Segmento:                       procCNJ.NomeSegmento,
		Grau:                           grauInferido,
		Classe:                         primeira.Classe,
		Assunto:                        assunto,
		OrgaoJulgador:                  primeira.Orgao,
		Comarca:                        comarca,
		Link:                           linkTribunal,
		LinkPeca:                       linkPecaMaisRecente,
		ValorCausa:                     valorCausa,
		ValorCausaFormatado:            datajud.FormatarDinheiroBR(valorCausa),
		DataAjuizamento:                dataAjuizamento,
		DataAjuizamentoFormatada:       dataAjuizamentoFormatada,
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

	pCNJ := cnj.Parse(proc.Numero)

	// 1. Identifica a peça mais recente entre as publicações do DJEN
	for _, pub := range pubs {
		linkP := pub.LinkPeca
		if linkP == "" && pub.Link != "" && (strings.Contains(pub.Link, "ConsultaDocumento") || strings.Contains(pub.Link, "listView.seam?x=")) {
			linkP = pub.Link
		}
		if linkP != "" {
			proc.LinkPeca = linkP
			break
		}
	}

	// 2. Garante que proc.Link aponte para a consulta oficial dos autos no Tribunal
	if proc.Link == "" || strings.Contains(proc.Link, "ConsultaDocumento") || strings.Contains(proc.Link, "listView.seam?x=") {
		proc.Link = cnj.GerarLinkTribunal(proc.Numero, proc.Tribunal, proc.Grau, pCNJ.OrigemCodigo)
	}

	// Extrai comarca exata da publicação ou do código de origem CNJ caso o DataJud não tenha informado (ou tenha vindo genérica da Turma Recursal)
	if proc.Comarca == "" || strings.Contains(proc.Comarca, ",") {
		for _, pub := range pubs {
			if c := comunica.InferirComarcaPorOrigem(proc.Tribunal, pCNJ.OrigemCodigo, pub.Orgao, pub.Texto); c != "" {
				proc.Comarca = c
				break
			}
		}
	}
	if proc.Comarca == "" {
		proc.Comarca = comunica.InferirComarcaPorOrigem(proc.Tribunal, pCNJ.OrigemCodigo, proc.OrgaoJulgador, "")
	}

	if proc.Grau == "" {
		proc.Grau = comunica.InferirGrau(proc.OrgaoJulgador, proc.Classe)
	}

	if proc.ValorCausa == nil {
		for _, pub := range pubs {
			if v := comunica.ExtrairValorCausaDoTexto(pub.Texto); v != nil {
				proc.ValorCausa = v
				proc.ValorCausaFormatado = datajud.FormatarDinheiroBR(v)
				break
			}
		}
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
			linkP := pub.LinkPeca
			if linkP == "" && pub.Link != "" && (strings.Contains(pub.Link, "ConsultaDocumento") || strings.Contains(pub.Link, "listView.seam?x=")) {
				linkP = pub.Link
			}
			proc.Movimentacoes = append(proc.Movimentacoes, domain.Movimentacao{
				Data:          t,
				DataFormatada: df,
				Tipo:          string(comunica.ClassificarTipo(pub.TipoComunicacao)),
				Descricao:     pub.TipoComunicacao,
				Complemento:   pub.Texto,
				Orgao:         pub.Orgao,
				Fonte:         string(domain.FonteComunica),
				Link:          pub.Link,
				LinkPeca:      linkP,
				IDExterno:     idExt,
			})
			jaTem[idExt] = true
		}
	}

	// Preenche órgão padrão, Link de consulta do Tribunal e correlaciona LinkPeca quando houver publicação DJEN correspondente
	for i := range proc.Movimentacoes {
		if proc.Movimentacoes[i].Orgao == "" && proc.OrgaoJulgador != "" {
			proc.Movimentacoes[i].Orgao = proc.OrgaoJulgador
		}
		if proc.Movimentacoes[i].Link == "" && proc.Link != "" {
			proc.Movimentacoes[i].Link = proc.Link
		}
		// Se for movimento do DataJud sem LinkPeca, verifica se é publicação/ato judicial correspondente a uma publicação do DJEN (janela de até 3 dias)
		if proc.Movimentacoes[i].Fonte == string(domain.FonteDataJud) && proc.Movimentacoes[i].LinkPeca == "" {
			tipoMov := proc.Movimentacoes[i].Tipo
			descLower := strings.ToLower(proc.Movimentacoes[i].Descricao)
			ehAtoPublicavel := tipoMov == string(domain.TipoPublicacaoDJE) ||
				tipoMov == string(domain.TipoIntimacao) ||
				tipoMov == string(domain.TipoSentenca) ||
				tipoMov == string(domain.TipoDecisao) ||
				tipoMov == string(domain.TipoDespacho) ||
				strings.Contains(descLower, "publica") ||
				strings.Contains(descLower, "diário") ||
				strings.Contains(descLower, "expedição")
			if ehAtoPublicavel && !proc.Movimentacoes[i].Data.IsZero() {
				for _, pub := range pubs {
					if pub.LinkPeca == "" {
						continue
					}
					tPub, _ := comunica.ParseDataDJEN(pub.DataDisponibilizacao)
					if !tPub.IsZero() {
						diff := tPub.Sub(proc.Movimentacoes[i].Data)
						if diff < 0 {
							diff = -diff
						}
						if diff <= 72*time.Hour {
							proc.Movimentacoes[i].LinkPeca = pub.LinkPeca
							break
						}
					}
				}
			}
		}
	}

	// Ordena movimentos do mais recente para o mais antigo
	sort.Slice(proc.Movimentacoes, func(i, j int) bool {
		return proc.Movimentacoes[i].Data.After(proc.Movimentacoes[j].Data)
	})

	if proc.DataAjuizamento == nil && len(proc.Movimentacoes) > 0 {
		maisAntiga := proc.Movimentacoes[len(proc.Movimentacoes)-1].Data
		if !maisAntiga.IsZero() {
			proc.DataAjuizamento = &maisAntiga
			proc.DataAjuizamentoFormatada = maisAntiga.Format("02/01/2006")
		}
	}

	// Enriquece partes e advogados do DataJud com o que veio do diário
	nomesPartes := make(map[string]bool)
	for _, p := range proc.Partes {
		nomesPartes[strings.ToUpper(strings.TrimSpace(p.Nome))] = true
	}
	for _, pub := range pubs {
		for _, parte := range pub.Partes {
			nomeLimpo := comunica.LimparNomeParte(parte.Nome)
			if !comunica.EhNomeParteValido(nomeLimpo) {
				continue
			}
			parte.Nome = nomeLimpo
			chave := strings.ToUpper(nomeLimpo)
			if !nomesPartes[chave] {
				nomesPartes[chave] = true
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

	type resExiste struct {
		existe bool
		fonte  string
	}
	ch := make(chan resExiste, 2)
	ctxBusca, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	go func() {
		pubs, err := o.buscarDJENComCache(ctxBusca, procCNJ.Digitos)
		ch <- resExiste{existe: err == nil && len(pubs) > 0, fonte: string(domain.FonteComunica)}
	}()

	go func() {
		p, err := o.buscarDataJudComCache(ctxBusca, procCNJ.Formatado, procCNJ.Digitos)
		ch <- resExiste{existe: err == nil && p != nil, fonte: string(domain.FonteDataJud)}
	}()

	for i := 0; i < 2; i++ {
		r := <-ch
		if r.existe {
			return true, r.fonte, nil
		}
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

	resumos := []domain.ResumoProcesso{}
	for num, lista := range agrupados {
		sort.Slice(lista, func(i, j int) bool {
			return lista[i].DataDisponibilizacao > lista[j].DataDisponibilizacao
		})
		primeira := lista[0]
		pCNJ := cnj.Parse(num)

		// Monta título com as partes se disponíveis nas publicações; caso contrário, usa a classe
		partesMap := make(map[string]domain.Parte)
		for _, pub := range lista {
			for _, parte := range pub.Partes {
				chave := strings.ToUpper(strings.TrimSpace(parte.Nome))
				if _, ok := partesMap[chave]; !ok || partesMap[chave].Polo == domain.PoloTerceiro {
					partesMap[chave] = parte
				}
			}
		}
		var partesProc []domain.Parte
		for _, p := range partesMap {
			partesProc = append(partesProc, p)
		}
		polosProc := domain.Polos{
			Ativo:   filtrarPolo(partesProc, domain.PoloAtivo),
			Passivo: filtrarPolo(partesProc, domain.PoloPassivo),
		}
		titulo := montarTitulo(polosProc)
		if titulo == "— × —" || strings.TrimSpace(titulo) == "" {
			titulo = primeira.Classe
		}

		resumos = append(resumos, domain.ResumoProcesso{
			Numero:                         num,
			Titulo:                         titulo,
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
	filtradas := []domain.Parte{}
	for _, p := range partes {
		if p.Polo == polo {
			filtradas = append(filtradas, p)
		}
	}
	return filtradas
}

func montarTitulo(polos domain.Polos) string {
	extrairNomes := func(partes []domain.Parte) []string {
		var nomes []string
		for _, p := range partes {
			n := comunica.LimparNomeParte(p.Nome)
			if !comunica.EhNomeParteValido(n) {
				continue
			}
			jaExiste := false
			for _, existente := range nomes {
				if strings.EqualFold(n, existente) {
					jaExiste = true
					break
				}
			}
			if !jaExiste {
				nomes = append(nomes, n)
			}
		}
		return nomes
	}

	ativos := extrairNomes(polos.Ativo)
	passivos := extrairNomes(polos.Passivo)

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
