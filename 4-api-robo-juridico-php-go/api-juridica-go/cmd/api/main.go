package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/configs"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/auth"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/cache"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/calculator"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/calculator/collectors"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/fontes/comunica"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/fontes/datajud"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/network"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/service"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/storage"
	transporthttp "github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/transport/http"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	cfg := configs.Carregar()

	log.Printf("[INICIALIZACAO] Iniciando Robô Jurídico API (Go) em %s:%s", cfg.Host, cfg.Porta)

	// 1. Rede e Proxy Pool
	var proxyPool *network.ProxyPool
	if len(cfg.Proxies) > 0 || cfg.GatewayHost != "" {
		proxyPool = network.NovoProxyPool(
			cfg.Proxies,
			cfg.GatewayHost,
			cfg.GatewayLogin,
			cfg.GatewaySenha,
			cfg.GatewayPortaInicial,
			cfg.GatewayPortas,
			4,
			16,
		)
		log.Printf("[INFO] Proxy Pool ativado com %d proxies e gateway configurado", len(cfg.Proxies))
	}

	httpCliente := network.NovoClienteHTTP(60*time.Second, cfg.UserAgentExtra)

	// 2. Clientes das Fontes Oficiais
	djClient := datajud.NovoClienteDataJud(httpCliente, cfg.DataJudAPIKey, cfg.DataJudTimeout, cfg.DataJudTentativas)
	comunicaClient := comunica.NovoClienteComunica(httpCliente, proxyPool, cfg.ComunicaTimeout, cfg.ComunicaTamanhoPagina, cfg.ComunicaTentativas)

	// 3. Cache em memória e Orquestrador de Fontes
	cacheTTL := cache.NovoCacheTTL()
	orquestrador := service.NovoOrquestrador(djClient, comunicaClient, cacheTTL)

	// 4. Banco de Dados SQLite (Painel, Chaves e Cotas)
	var db storage.BancoDados
	if cfg.BancoCaminho != "" {
		dir := filepath.Dir(cfg.BancoCaminho)
		if dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0755)
		}

		sqliteDB, err := storage.NovoBancoSQLite(cfg.BancoCaminho)
		if err != nil {
			log.Printf("[AVISO] Nao foi possivel conectar ao SQLite em '%s': %v. Operando em modo somente-memoria/chaves internas.", cfg.BancoCaminho, err)
		} else {
			db = sqliteDB
			defer func() {
				if err := db.Fechar(); err != nil {
					log.Printf("[ERRO] Erro ao fechar SQLite: %v", err)
				}
			}()
			log.Printf("[INFO] SQLite conectado com sucesso: %s", cfg.BancoCaminho)
		}
	}

	// 5. Servico de Autenticacao
	servicoAuth := auth.NovoServicoAuth(db, cfg.ChavesInternas, cfg.LimitePadraoRPM)
	log.Printf("[INFO] Servico de autenticacao pronto com %d chaves internas configuradas", len(cfg.ChavesInternas))

	// 6. Motor de Cálculos Judiciais e Índices Econômicos
	calcRepo := calculator.NovoRepositorioIndices()
	calcEngine := calculator.NovoEngineCalculo(calcRepo)

	// Worker Agendado Diário / Mensal (Bacen, IBGE, Tribunais)
	workerColetor := collectors.NovoWorker(calcRepo, 24*time.Hour)
	ctxWorker, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	workerColetor.Iniciar(ctxWorker)

	log.Printf("[INFO] Motor de calculos judiciais ativado com %d indices e worker agendado", len(calcRepo.ListarIndices()))

	// 7. Camada HTTP e Rotas
	handlers := transporthttp.NovosHandlers(orquestrador, db, "1.0.0", calcEngine)
	handlers.SetWorkerColetor(workerColetor)
	router := transporthttp.NovoRoteador(handlers, servicoAuth, db)

	endereco := fmt.Sprintf("%s:%s", cfg.Host, cfg.Porta)
	srv := &http.Server{
		Addr:         endereco,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 7. Encerramento gracioso (Graceful Shutdown)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[PRONTO] Servidor HTTP escutando em http://%s", endereco)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] Falha no servidor HTTP: %v", err)
		}
	}()

	<-stop
	log.Println("[SHUTDOWN] Sinal de parada recebido. Encerrando conexoes ativas...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("[ERRO] Falha no encerramento gracioso: %v", err)
	}

	log.Println("[SUCESSO] Robô Jurídico API encerrado com sucesso.")
}
