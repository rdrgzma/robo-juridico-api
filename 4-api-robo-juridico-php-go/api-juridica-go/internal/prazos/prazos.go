package prazos

import (
	"fmt"
	"strings"
	"time"
)

// CalculoPrazoRequest define os parâmetros de entrada para apuração de prazo processual fatal
type CalculoPrazoRequest struct {
	DataDisponibilizacao string   `json:"data_disponibilizacao"`        // AAAA-MM-DD ou DD/MM/AAAA (Data no DJe/DJEN ou data da intimação)
	DiasPrazo            int      `json:"dias_prazo"`                   // Ex: 5, 8, 10, 15
	Rito                 string   `json:"rito,omitempty"`               // "CPC_UTEIS" (padrão), "CLT_UTEIS", "CORRIDOS" (Penal/JEC/MS)
	TipoIntimacao        string   `json:"tipo_intimacao,omitempty"`     // "DIARIO_ELETRONICO" (padrão: publica no 1º dia útil seguinte) ou "DIRETA" (citação/mandado juntado)
	PrazoEmDobro         bool     `json:"prazo_em_dobro,omitempty"`     // Fazenda Pública (Art. 183 CPC), MP (Art. 180), Defensoria (Art. 186)
	SuspenderRecesso     *bool    `json:"suspender_recesso,omitempty"`  // Art. 220 CPC (20/dez a 20/jan) - padrão true para rito CPC/CLT
	FeriadosLocais       []string `json:"feriados_locais,omitempty"`    // Datas adicionais AAAA-MM-DD ou DD/MM/AAAA (feriados municipais/estaduais/indisponibilidade)
	Tribunal             string   `json:"tribunal,omitempty"`           // Ex: "TJSP", "TJRJ", "TRF1"
}

// DiaEvolucaoPrazo detalha cada dia percorrido na contagem para auditoria do advogado
type DiaEvolucaoPrazo struct {
	Data       string `json:"data"`
	DiaSemana  string `json:"dia_semana"`
	Util       bool   `json:"util"`
	DiaContado int    `json:"dia_contado"`
	Motivo     string `json:"motivo"`
}

// CalculoPrazoResponse retorna o resultado pericial do prazo processual
type CalculoPrazoResponse struct {
	Sucesso              bool               `json:"sucesso"`
	Rito                 string             `json:"rito"`
	DiasPrazoOriginal    int                `json:"dias_prazo_original"`
	DiasPrazoEfetivo     int                `json:"dias_prazo_efetivo"`
	DataDisponibilizacao string             `json:"data_disponibilizacao"`
	DataPublicacao       string             `json:"data_publicacao"`
	DataInicioContagem   string             `json:"data_inicio_contagem"`
	DataVencimentoFatal  string             `json:"data_vencimento_fatal"`
	FundamentacaoLegal   []string           `json:"fundamentacao_legal"`
	MemoriaContagem      []DiaEvolucaoPrazo `json:"memoria_contagem"`
}

// FeriadoForense representa um feriado nacional ou forense
type FeriadoForense struct {
	Data      string `json:"data"`
	Descricao string `json:"descricao"`
	Tipo      string `json:"tipo"`
}

// CalcularPrazoProcessual executa a contagem oficial conforme Art. 224 e Art. 220 do CPC/2015 ou Art. 775 da CLT
func CalcularPrazoProcessual(req CalculoPrazoRequest) (*CalculoPrazoResponse, error) {
	if req.DiasPrazo <= 0 || req.DiasPrazo > 365 {
		return nil, fmt.Errorf("dias_prazo deve estar entre 1 e 365")
	}

	dtDisp, err := parseData(req.DataDisponibilizacao)
	if err != nil {
		return nil, fmt.Errorf("data_disponibilizacao invalida ('%s'): use AAAA-MM-DD ou DD/MM/AAAA", req.DataDisponibilizacao)
	}

	rito := strings.ToUpper(strings.TrimSpace(req.Rito))
	if rito == "" {
		rito = "CPC_UTEIS"
	}
	apenasDiasUteis := rito != "CORRIDOS"

	suspenderRecesso := true
	if req.SuspenderRecesso != nil {
		suspenderRecesso = *req.SuspenderRecesso
	} else if rito == "CORRIDOS" {
		suspenderRecesso = false
	}

	// Mapa de feriados customizados/locais
	feriadosCustom := make(map[string]string)
	for _, fStr := range req.FeriadosLocais {
		if dtF, err := parseData(fStr); err == nil {
			feriadosCustom[dtF.Format("2006-01-02")] = "Feriado Local / Suspensão de Expediente"
		}
	}

	diasEfetivos := req.DiasPrazo
	if req.PrazoEmDobro {
		diasEfetivos *= 2
	}

	tipoIntimacao := strings.ToUpper(strings.TrimSpace(req.TipoIntimacao))
	if tipoIntimacao == "" {
		tipoIntimacao = "DIARIO_ELETRONICO"
	}

	// 1. Determina a Data de Publicação (Art. 224, § 2º CPC: primeiro dia útil seguinte à disponibilização no DJe)
	dtPublicacao := dtDisp
	if tipoIntimacao == "DIARIO_ELETRONICO" {
		dtPublicacao = proximoDiaUtil(dtDisp.AddDate(0, 0, 1), suspenderRecesso, feriadosCustom)
	} else {
		// Intimação direta: se cair em dia não útil, prorroga para o primeiro dia útil seguinte
		dtPublicacao = proximoDiaUtil(dtDisp, suspenderRecesso, feriadosCustom)
	}

	// 2. Determina o Termo Inicial da Contagem (Art. 224, § 3º CPC: primeiro dia útil seguinte ao da publicação)
	dtInicio := proximoDiaUtil(dtPublicacao.AddDate(0, 0, 1), suspenderRecesso, feriadosCustom)

	// 3. Percorre os dias até completar diasEfetivos
	cursor := dtInicio
	contados := 0
	var memoria []DiaEvolucaoPrazo
	var dtVencimento time.Time

	for contados < diasEfetivos {
		util, motivo := verificarDiaUtil(cursor, suspenderRecesso, feriadosCustom)

		if apenasDiasUteis {
			if util {
				contados++
				motivo = fmt.Sprintf("%dº dia útil do prazo", contados)
			}
		} else {
			// Prazo em dias corridos: conta todos os dias, mas se o último dia cair em feriado/fim de semana, prorroga
			if contados+1 == diasEfetivos && !util {
				motivo = "Vencimento prorrogado (dia não útil - Art. 224 § 1º CPC)"
			} else {
				contados++
				motivo = fmt.Sprintf("%dº dia corrido do prazo", contados)
			}
		}

		memoria = append(memoria, DiaEvolucaoPrazo{
			Data:       cursor.Format("02/01/2006"),
			DiaSemana:  nomeDiaSemana(cursor.Weekday()),
			Util:       util,
			DiaContado: contados,
			Motivo:     motivo,
		})

		if contados == diasEfetivos {
			dtVencimento = cursor
			break
		}
		cursor = cursor.AddDate(0, 0, 1)
	}

	fundamentacao := []string{
		"Art. 224 do CPC/2015: Exclusão do dia do começo e inclusão do dia do vencimento.",
	}
	if tipoIntimacao == "DIARIO_ELETRONICO" {
		fundamentacao = append(fundamentacao, "Art. 224, § 2º e § 3º do CPC/2015 (Lei 11.419/2006, Art. 4º, § 3º e § 4º): Considera-se como data de publicação o primeiro dia útil seguinte ao da disponibilização no Diário da Justiça eletrônico, iniciando-se a contagem no primeiro dia útil subsequente.")
	}
	if apenasDiasUteis {
		fundamentacao = append(fundamentacao, "Art. 219 do CPC/2015 / Art. 775 da CLT: Na contagem de prazo em dias, computar-se-ão somente os dias úteis.")
	}
	if suspenderRecesso {
		fundamentacao = append(fundamentacao, "Art. 220 do CPC/2015: Suspensão do curso do prazo processual nos dias compreendidos entre 20 de dezembro e 20 de janeiro (Recesso Forense).")
	}
	if req.PrazoEmDobro {
		fundamentacao = append(fundamentacao, "Prazo computado em dobro nos termos dos Arts. 180 (MP), 183 (Fazenda Pública) ou 186 (Defensoria Pública) do CPC/2015.")
	}

	return &CalculoPrazoResponse{
		Sucesso:              true,
		Rito:                 rito,
		DiasPrazoOriginal:    req.DiasPrazo,
		DiasPrazoEfetivo:     diasEfetivos,
		DataDisponibilizacao: dtDisp.Format("02/01/2006"),
		DataPublicacao:       dtPublicacao.Format("02/01/2006"),
		DataInicioContagem:   dtInicio.Format("02/01/2006"),
		DataVencimentoFatal:  dtVencimento.Format("02/01/2006"),
		FundamentacaoLegal:   fundamentacao,
		MemoriaContagem:      memoria,
	}, nil
}

// ListarFeriadosAno retorna a lista completa de feriados nacionais e forenses para o ano solicitado
func ListarFeriadosAno(ano int) []FeriadoForense {
	if ano < 1990 || ano > 2100 {
		ano = time.Now().Year()
	}

	pascoa := calcularDomingoPascoa(ano)
	carnavalSeg := pascoa.AddDate(0, 0, -48)
	carnavalTer := pascoa.AddDate(0, 0, -47)
	quartaCinzas := pascoa.AddDate(0, 0, -46)
	quintaSanta := pascoa.AddDate(0, 0, -3)
	sextaPaixao := pascoa.AddDate(0, 0, -2)
	corpusChristi := pascoa.AddDate(0, 0, 60)

	return []FeriadoForense{
		{Data: fmt.Sprintf("%04d-01-01", ano), Descricao: "Confraternização Universal (Lei 662/1949)", Tipo: "NACIONAL"},
		{Data: carnavalSeg.Format("2006-01-02"), Descricao: "Segunda-feira de Carnaval (Lei 5.010/1966)", Tipo: "FORENSE"},
		{Data: carnavalTer.Format("2006-01-02"), Descricao: "Terça-feira de Carnaval (Lei 5.010/1966)", Tipo: "FORENSE"},
		{Data: quartaCinzas.Format("2006-01-02"), Descricao: "Quarta-feira de Cinzas (Ponto Facultativo Forense)", Tipo: "FORENSE"},
		{Data: quintaSanta.Format("2006-01-02"), Descricao: "Quinta-feira Santa (Lei 5.010/1966 - Art. 62, II)", Tipo: "FORENSE"},
		{Data: sextaPaixao.Format("2006-01-02"), Descricao: "Sexta-feira da Paixão (Lei 9.093/1995)", Tipo: "NACIONAL"},
		{Data: fmt.Sprintf("%04d-04-21", ano), Descricao: "Tiradentes (Lei 662/1949)", Tipo: "NACIONAL"},
		{Data: fmt.Sprintf("%04d-05-01", ano), Descricao: "Dia Mundial do Trabalho (Lei 662/1949)", Tipo: "NACIONAL"},
		{Data: corpusChristi.Format("2006-01-02"), Descricao: "Corpus Christi", Tipo: "FORENSE"},
		{Data: fmt.Sprintf("%04d-08-11", ano), Descricao: "Dia do Magistrado e do Advogado - Pendura (Lei 5.010/1966)", Tipo: "FORENSE"},
		{Data: fmt.Sprintf("%04d-09-07", ano), Descricao: "Independência do Brasil (Lei 662/1949)", Tipo: "NACIONAL"},
		{Data: fmt.Sprintf("%04d-10-12", ano), Descricao: "Nossa Senhora Aparecida (Lei 6.802/1980)", Tipo: "NACIONAL"},
		{Data: fmt.Sprintf("%04d-10-28", ano), Descricao: "Dia do Servidor Público (Art. 236 Lei 8.112/1990)", Tipo: "FORENSE"},
		{Data: fmt.Sprintf("%04d-11-01", ano), Descricao: "Dia de Todos os Santos (Lei 5.010/1966 - Art. 62, IV)", Tipo: "FORENSE"},
		{Data: fmt.Sprintf("%04d-11-02", ano), Descricao: "Finados (Lei 10.607/2002)", Tipo: "NACIONAL"},
		{Data: fmt.Sprintf("%04d-11-15", ano), Descricao: "Proclamação da República (Lei 662/1949)", Tipo: "NACIONAL"},
		{Data: fmt.Sprintf("%04d-11-20", ano), Descricao: "Dia Nacional de Zumbi e da Consciência Negra (Lei 14.759/2023)", Tipo: "NACIONAL"},
		{Data: fmt.Sprintf("%04d-12-08", ano), Descricao: "Dia da Justiça (Decreto-Lei 8.292/1945 e Lei 5.010/1966)", Tipo: "FORENSE"},
		{Data: fmt.Sprintf("%04d-12-25", ano), Descricao: "Natal (Lei 662/1949)", Tipo: "NACIONAL"},
	}
}

func proximoDiaUtil(dt time.Time, suspenderRecesso bool, feriadosCustom map[string]string) time.Time {
	cursor := dt
	for {
		util, _ := verificarDiaUtil(cursor, suspenderRecesso, feriadosCustom)
		if util {
			return cursor
		}
		cursor = cursor.AddDate(0, 0, 1)
	}
}

func verificarDiaUtil(dt time.Time, suspenderRecesso bool, feriadosCustom map[string]string) (bool, string) {
	if dt.Weekday() == time.Saturday {
		return false, "Sábado (sem expediente forense)"
	}
	if dt.Weekday() == time.Sunday {
		return false, "Domingo (sem expediente forense)"
	}

	if suspenderRecesso {
		if (dt.Month() == time.December && dt.Day() >= 20) || (dt.Month() == time.January && dt.Day() <= 20) {
			return false, "Recesso Forense / Suspensão de Prazos (Art. 220 do CPC - 20/dez a 20/jan)"
		}
	}

	chave := dt.Format("2006-01-02")
	if desc, ok := feriadosCustom[chave]; ok {
		return false, desc
	}

	for _, f := range ListarFeriadosAno(dt.Year()) {
		if f.Data == chave {
			return false, f.Descricao
		}
	}

	return true, "Dia útil forense"
}

// Algoritmo de Meeus/Jones/Butcher para cálculo exato do Domingo de Páscoa
func calcularDomingoPascoa(ano int) time.Time {
	a := ano % 19
	b := ano / 100
	c := ano % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	mes := (h + l - 7*m + 114) / 31
	dia := ((h + l - 7*m + 114) % 31) + 1
	return time.Date(ano, time.Month(mes), dia, 0, 0, 0, 0, time.UTC)
}

func parseData(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"2006-01-02", "02/01/2006"} {
		if dt, err := time.Parse(layout, s); err == nil {
			return time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, time.UTC), nil
		}
	}
	return time.Time{}, fmt.Errorf("formato de data invalido")
}

func nomeDiaSemana(w time.Weekday) string {
	switch w {
	case time.Sunday:
		return "Domingo"
	case time.Monday:
		return "Segunda-feira"
	case time.Tuesday:
		return "Terça-feira"
	case time.Wednesday:
		return "Quarta-feira"
	case time.Thursday:
		return "Quinta-feira"
	case time.Friday:
		return "Sexta-feira"
	case time.Saturday:
		return "Sábado"
	default:
		return ""
	}
}
