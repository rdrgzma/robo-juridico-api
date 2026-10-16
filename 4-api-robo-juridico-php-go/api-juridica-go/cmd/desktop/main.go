package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/bridge"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/licensing"
)

func main() {
	comando := flag.String("cmd", "info", "Comando a executar: info | gerar-chaves | emitir-trial | fingerprint")
	flag.Parse()

	switch *comando {
	case "fingerprint":
		hwID := licensing.ObterHardwareFingerprint()
		fmt.Printf("Hardware Fingerprint do Cliente: %s\n", hwID)

	case "gerar-chaves":
		_, privPEM, pubPEM, err := licensing.GerarChavesECDSA()
		if err != nil {
			log.Fatalf("Erro ao gerar chaves: %v", err)
		}
		fmt.Println("=== CHAVE PRIVADA ECDSA (GUARDE EM SEGREDO NO SERVIDOR) ===")
		fmt.Println(privPEM)
		fmt.Println("=== CHAVE PUBLICA ECDSA (EMBUTA NO APLICATIVO DESKTOP) ===")
		fmt.Println(pubPEM)

	case "info":
		cfg := bridge.ConfigDesktop{
			Versao: "1.0.0-desktop",
		}
		core, err := bridge.NovoCoreBridge(cfg)
		if err != nil {
			log.Fatalf("Erro ao iniciar core: %v", err)
		}
		info := core.ObterInfoSistema()
		out, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println("=== INFORMACOES DO SISTEMA DESKTOP ===")
		fmt.Println(string(out))

	default:
		fmt.Printf("Comando desconhecido: %s. Use -cmd=info|gerar-chaves|fingerprint\n", *comando)
		os.Exit(1)
	}
}
