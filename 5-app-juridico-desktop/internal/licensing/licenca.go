package licensing

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var (
	ErrLicencaInvalida   = errors.New("licenca invalida ou formato corrompido")
	ErrAssinaturaInvalida = errors.New("assinatura digital da licenca nao confere com a chave oficial")
	ErrLicencaExpirada   = errors.New("a licenca de uso expirou")
	ErrHardwareInvalido  = errors.New("esta licenca pertence a outro computador (hardware ID divergente)")
	ErrRecursoNaoAutorizado = errors.New("recurso nao autorizado para este plano de licenca")
)

type RecursosLicenca struct {
	CalculoJudicialCompleto bool `json:"calculo_judicial_completo"` // Lei 14.905/2024, Art. 354 CC
	RevisaoBancaria         bool `json:"revisao_bancaria"`          // Price, SAC, Gauss
	ContagemPrazos          bool `json:"contagem_prazos"`           // CPC e CLT
	ConsultaProcessual      bool `json:"consulta_processual"`       // DataJud e DJEN
	ExportarExcel           bool `json:"exportar_excel"`            // .xlsx
	ExportarLaudoPDF        bool `json:"exportar_laudo_pdf"`        // Laudo com QR Code e Hash
	AtualizacaoAutomatica   bool `json:"atualizacao_automatica"`    // Sincronização automática Bacen/IBGE
}

type DadosLicenca struct {
	IDCliente       string          `json:"id_cliente"`
	NomeTitular     string          `json:"nome_titular"`
	Documento       string          `json:"documento"`        // CPF ou CNPJ
	Plano           string          `json:"plano"`            // "trial", "basico", "pericial_pro", "ilimitado"
	HardwareID      string          `json:"hardware_id"`      // Se vazio, não amarra à máquina
	EmitidaEm       time.Time       `json:"emitida_em"`
	ValidaAte       time.Time       `json:"valida_ate"`
	LimiteCalculos  int             `json:"limite_calculos"`  // 0 para ilimitado
	LimiteConsultas int             `json:"limite_consultas"` // 0 para ilimitado
	Recursos        RecursosLicenca `json:"recursos"`
}

type PacoteLicenca struct {
	PayloadB64    string `json:"payload"`
	AssinaturaB64 string `json:"assinatura"`
}

type ecdsaSignature struct {
	R, S *big.Int
}

// ChavePúblicaPadrãoEmbutida para validação local (substituível em produção)
// Usamos curva P-256 (NIST P-256 / secp256r1)
type ValidadorLicenca struct {
	publicKey *ecdsa.PublicKey
}

func NovoValidador(publicKeyPEM string) (*ValidadorLicenca, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("falha ao decodificar PEM da chave publica")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("falha ao parsear chave publica: %w", err)
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("chave publica fornecida nao e do tipo ECDSA")
	}

	return &ValidadorLicenca{publicKey: ecdsaPub}, nil
}

// Validar decodifica, verifica a assinatura criptográfica e as restrições da licença.
func (v *ValidadorLicenca) Validar(tokenLicenca string, hardwareIDAtual string) (*DadosLicenca, error) {
	tokenLimpo := strings.TrimSpace(tokenLicenca)
	if tokenLimpo == "" {
		return nil, ErrLicencaInvalida
	}

	var pacote PacoteLicenca
	// Tenta JSON direto ou Base64 do pacote
	err := json.Unmarshal([]byte(tokenLimpo), &pacote)
	if err != nil {
		// Tenta decodificar de base64 primeiro
		raw, errB64 := base64.StdEncoding.DecodeString(tokenLimpo)
		if errB64 != nil {
			return nil, ErrLicencaInvalida
		}
		if errJSON := json.Unmarshal(raw, &pacote); errJSON != nil {
			return nil, ErrLicencaInvalida
		}
	}

	// 1. Validar Assinatura Digital
	digest := sha256.Sum256([]byte(pacote.PayloadB64))
	sigBytes, err := base64.StdEncoding.DecodeString(pacote.AssinaturaB64)
	if err != nil {
		return nil, ErrAssinaturaInvalida
	}

	var sig ecdsaSignature
	_, err = asn1.Unmarshal(sigBytes, &sig)
	if err != nil {
		return nil, ErrAssinaturaInvalida
	}

	if !ecdsa.Verify(v.publicKey, digest[:], sig.R, sig.S) {
		return nil, ErrAssinaturaInvalida
	}

	// 2. Decodificar Payload
	payloadBytes, err := base64.StdEncoding.DecodeString(pacote.PayloadB64)
	if err != nil {
		return nil, ErrLicencaInvalida
	}

	var dados DadosLicenca
	if err := json.Unmarshal(payloadBytes, &dados); err != nil {
		return nil, ErrLicencaInvalida
	}

	// 3. Validar Validade Temporal
	if !dados.ValidaAte.IsZero() && time.Now().After(dados.ValidaAte) {
		return &dados, ErrLicencaExpirada
	}

	// 4. Validar Hardware ID se especificado na licença
	if dados.HardwareID != "" && hardwareIDAtual != "" {
		if !strings.EqualFold(dados.HardwareID, hardwareIDAtual) {
			return &dados, ErrHardwareInvalido
		}
	}

	return &dados, nil
}

// EmissorLicenca (utilizado pelo servidor ou pelo gerador de chaves para emitir licenças assinadas)
type EmissorLicenca struct {
	privateKey *ecdsa.PrivateKey
}

func NovoEmissor(privateKey *ecdsa.PrivateKey) *EmissorLicenca {
	return &EmissorLicenca{privateKey: privateKey}
}

func GerarChavesECDSA() (*ecdsa.PrivateKey, string, string, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, "", "", err
	}

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, "", "", err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, "", "", err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return priv, string(privPEM), string(pubPEM), nil
}

func (e *EmissorLicenca) Emitir(dados DadosLicenca) (string, error) {
	payloadBytes, err := json.Marshal(dados)
	if err != nil {
		return "", err
	}

	payloadB64 := base64.StdEncoding.EncodeToString(payloadBytes)
	digest := sha256.Sum256([]byte(payloadB64))

	r, s, err := ecdsa.Sign(rand.Reader, e.privateKey, digest[:])
	if err != nil {
		return "", err
	}

	sigBytes, err := asn1.Marshal(ecdsaSignature{R: r, S: s})
	if err != nil {
		return "", err
	}

	pacote := PacoteLicenca{
		PayloadB64:    payloadB64,
		AssinaturaB64: base64.StdEncoding.EncodeToString(sigBytes),
	}

	pacoteBytes, err := json.Marshal(pacote)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(pacoteBytes), nil
}
