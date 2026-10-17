package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/rdrgzma/robo-juridico-desktop/internal/bridge"
	"github.com/rdrgzma/robo-juridico-desktop/internal/desktopserver"
	"github.com/rdrgzma/robo-juridico-desktop/internal/licensing"
	"github.com/rdrgzma/robo-juridico-desktop/ui"
)

const VersaoApp = "1.0.0"

func main() {
	porta := flag.String("porta", "7890", "Porta local para a interface desktop")
	semNavegador := flag.Bool("no-browser", false, "Nao abrir o navegador automaticamente ao iniciar")
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime)

	hwID := licensing.ObterHardwareFingerprint()

	// Diretório de dados local do usuário (ex: dados/desktop.db)
	appDataDir := "dados"
	_ = os.MkdirAll(appDataDir, 0755)
	dbPath := filepath.Join(appDataDir, "desktop.db")

	dataJudKey := os.Getenv("DATAJUD_API_KEY")
	if dataJudKey == "" {
		dataJudKey = "cDZHYzlZa0JadVREZDJCendQbXY6SkJlTzNjLV9TRENyQk1RdnFKZGRQdw=="
	}

	cfg := bridge.ConfigDesktop{
		Versao:        VersaoApp,
		DataJudAPIKey: dataJudKey,
	}

	core, err := bridge.NovoCoreBridge(cfg)
	if err != nil {
		log.Fatalf("[FATAL] Erro ao iniciar motor juridico: %v", err)
	}

	server, err := desktopserver.NovoDesktopServer(core, ui.ObterFS(), dbPath)
	if err != nil {
		log.Fatalf("[FATAL] Erro ao iniciar servidor desktop: %v", err)
	}
	defer server.Fechar()

	endereco := fmt.Sprintf("127.0.0.1:%s", *porta)
	listener, err := net.Listen("tcp", endereco)
	if err != nil {
		// Se a porta 7890 estiver ocupada, tenta porta aleatória disponível
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatalf("[FATAL] Nao foi possivel abrir porta de rede local: %v", err)
		}
	}
	portaReal := listener.Addr().(*net.TCPAddr).Port
	urlApp := fmt.Sprintf("http://127.0.0.1:%d", portaReal)

	httpServer := &http.Server{
		Handler:      server.Router(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	fmt.Println("==================================================================")
	fmt.Printf("⚖️   ROBO JURIDICO DESKTOP ON-PREMISE (v%s)\n", VersaoApp)
	fmt.Printf("💻  Sistema: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("🔑  Hardware Fingerprint: %s\n", hwID)
	fmt.Printf("🌐  Interface Local: %s\n", urlApp)
	fmt.Println("==================================================================")
	fmt.Println("Pressione Ctrl+C para encerrar o aplicativo.")
	fmt.Println()

	// Inicia servidor em background
	go func() {
		if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[ERRO] Falha no servidor HTTP desktop: %v", err)
		}
	}()

	// Abre o navegador padrão na máquina do cliente
	if !*semNavegador {
		go func() {
			time.Sleep(300 * time.Millisecond)
			abrirNavegador(urlApp)
		}()
	}

	// Aguarda sinal de encerramento
	parar := make(chan os.Signal, 1)
	signal.Notify(parar, os.Interrupt, syscall.SIGTERM)
	<-parar

	fmt.Println("\n[INFO] Encerrando o Robo Juridico Desktop...")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctxShutdown)
	fmt.Println("[INFO] Aplicativo finalizado com sucesso.")
}

func abrirNavegador(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return
	}

	_ = cmd.Start()
}
