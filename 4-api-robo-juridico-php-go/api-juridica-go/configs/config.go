package configs

import (
	"crypto/sha256"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host                   string
	Porta                  string
	DataJudAPIKey          string
	DataJudTimeout         time.Duration
	DataJudTentativas      int
	ComunicaTimeout        time.Duration
	ComunicaTamanhoPagina  int
	ComunicaTentativas     int
	BancoCaminho           string
	Proxies                []string
	GatewayHost            string
	GatewayLogin           string
	GatewaySenha           string
	GatewayPortaInicial    int
	GatewayPortas          int
	UserAgentExtra         string
	LimitePadraoRPM        int
	ChavesInternas         map[string][32]byte
}

func Carregar() *Config {
	return &Config{
		Host:                  env("HOST", "0.0.0.0"),
		Porta:                 env("PORT", "8000"),
		DataJudAPIKey:         env("DATAJUD_API_KEY", "cDZHYzlZa0JadVREZDJCendQbXY6SkJlTzNjLV9TRENyQk1RdnFKZGRQdw=="),
		DataJudTimeout:        time.Duration(envInt("DATAJUD_TIMEOUT", 90)) * time.Second,
		DataJudTentativas:     envInt("DATAJUD_TENTATIVAS", 3),
		ComunicaTimeout:       time.Duration(envInt("COMUNICA_TIMEOUT", 12)) * time.Second,
		ComunicaTamanhoPagina: envInt("COMUNICA_TAMANHO_PAGINA", 100),
		ComunicaTentativas:    envInt("COMUNICA_TENTATIVAS", 3),
		BancoCaminho:          env("API_BANCO", "dados/api.db"),
		Proxies:               parseLista(env("CONSULTA_PROXIES", "")),
		GatewayHost:           env("CONSULTA_GATEWAY_HOST", ""),
		GatewayLogin:          env("CONSULTA_GATEWAY_LOGIN", ""),
		GatewaySenha:          env("CONSULTA_GATEWAY_SENHA", ""),
		GatewayPortaInicial:   envInt("CONSULTA_GATEWAY_PORTA_INICIAL", 10000),
		GatewayPortas:         envInt("CONSULTA_GATEWAY_PORTAS", 0),
		UserAgentExtra:        env("CONSULTA_USER_AGENT_EXTRA", ""),
		LimitePadraoRPM:       envInt("API_LIMITE_POR_MINUTO", 60),
		ChavesInternas:        parseChavesInternas(env("API_CHAVES", "")),
	}
}

func env(chave, padrao string) string {
	if val := os.Getenv(chave); val != "" {
		return strings.TrimSpace(val)
	}
	return padrao
}

func envInt(chave string, padrao int) int {
	if val := os.Getenv(chave); val != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			return n
		}
	}
	return padrao
}

func parseLista(texto string) []string {
	var res []string
	for _, item := range strings.Split(strings.ReplaceAll(texto, ";", ","), ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			res = append(res, item)
		}
	}
	return res
}

func parseChavesInternas(texto string) map[string][32]byte {
	chaves := make(map[string][32]byte)
	for _, item := range strings.Split(texto, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		partes := strings.SplitN(item, ":", 2)
		if len(partes) == 2 && len(partes[1]) >= 32 {
			nome := strings.TrimSpace(partes[0])
			chave := strings.TrimSpace(partes[1])
			chaves[nome] = sha256.Sum256([]byte(chave))
		}
	}
	return chaves
}
