package licensing

import (
	"strings"
	"testing"
	"time"
)

func TestLicenciamentoEHardwareFingerprint(t *testing.T) {
	// 1. Teste de Hardware Fingerprint
	hwID := ObterHardwareFingerprint()
	if hwID == "" {
		t.Fatalf("Esperava hardware fingerprint valido, obteve vazio")
	}
	if !strings.HasPrefix(hwID, "JUR-") {
		t.Fatalf("Hardware fingerprint deve iniciar com prefixo 'JUR-', obteve: %s", hwID)
	}

	// 2. Geração de Par de Chaves Criptográficas
	privKey, _, pubPEM, err := GerarChavesECDSA()
	if err != nil {
		t.Fatalf("Erro ao gerar chaves ECDSA: %v", err)
	}

	emissor := NovoEmissor(privKey)
	validador, err := NovoValidador(pubPEM)
	if err != nil {
		t.Fatalf("Erro ao inicializar validador com chave publica: %v", err)
	}

	// 3. Emissão de Licença Válida
	dadosOriginal := DadosLicenca{
		IDCliente:       "cli_teste_123",
		NomeTitular:     "Dr. Advogado Perito",
		Documento:       "123.456.789-00",
		Plano:           "pericial_pro",
		HardwareID:      hwID,
		EmitidaEm:       time.Now(),
		ValidaAte:       time.Now().Add(30 * 24 * time.Hour), // 30 dias
		LimiteCalculos:  500,
		LimiteConsultas: 200,
		Recursos: RecursosLicenca{
			CalculoJudicialCompleto: true,
			RevisaoBancaria:         true,
			ContagemPrazos:          true,
			ConsultaProcessual:      true,
			ExportarExcel:           true,
			ExportarLaudoPDF:        true,
			AtualizacaoAutomatica:   true,
		},
	}

	tokenLicenca, err := emissor.Emitir(dadosOriginal)
	if err != nil {
		t.Fatalf("Erro ao emitir token de licenca: %v", err)
	}

	// 4. Validação Positiva
	dadosValidados, err := validador.Validar(tokenLicenca, hwID)
	if err != nil {
		t.Fatalf("Falha inesperada ao validar licenca valida: %v", err)
	}

	if dadosValidados.NomeTitular != dadosOriginal.NomeTitular {
		t.Errorf("Nome titular divergente: %s != %s", dadosValidados.NomeTitular, dadosOriginal.NomeTitular)
	}
	if !dadosValidados.Recursos.CalculoJudicialCompleto {
		t.Errorf("Esperava CalculoJudicialCompleto = true")
	}

	// 5. Teste de Hardware Divergente
	_, errHwMismatch := validador.Validar(tokenLicenca, "JUR-OUTRA-MAQUINA-9999")
	if errHwMismatch != ErrHardwareInvalido {
		t.Errorf("Esperava ErrHardwareInvalido para maquina diferente, obteve: %v", errHwMismatch)
	}

	// 6. Teste de Licença Expirada
	dadosExpirados := dadosOriginal
	dadosExpirados.ValidaAte = time.Now().Add(-24 * time.Hour)
	tokenExpirado, err := emissor.Emitir(dadosExpirados)
	if err != nil {
		t.Fatalf("Erro ao emitir licenca expirada: %v", err)
	}

	_, errExp := validador.Validar(tokenExpirado, hwID)
	if errExp != ErrLicencaExpirada {
		t.Errorf("Esperava ErrLicencaExpirada, obteve: %v", errExp)
	}

	// 7. Teste de Adulteração (Tampering)
	tokenAdulterado := tokenLicenca[:len(tokenLicenca)-10] + "AAAAAAAAAA"
	_, errTamper := validador.Validar(tokenAdulterado, hwID)
	if errTamper == nil {
		t.Errorf("Esperava erro ao validar token adulterado, mas passou!")
	}
}
