package licensing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ObterHardwareFingerprint gera um identificador único e estável da máquina do cliente.
// No Windows: Consulta o UUID da BIOS/Placa-mãe ou MachineGuid.
// No Linux: Lê /sys/class/dmi/id/product_uuid ou /etc/machine-id.
// Fallback: Combinação de MAC address das interfaces de rede + Hostname.
func ObterHardwareFingerprint() string {
	rawID := ""

	switch runtime.GOOS {
	case "windows":
		rawID = obterUUIDWindows()
	case "linux":
		rawID = obterUUIDLinux()
	default:
		rawID = obterFallbackHardware()
	}

	if rawID == "" {
		rawID = obterFallbackHardware()
	}

	hash := sha256.Sum256([]byte("ROBO_JURIDICO_HW:" + strings.TrimSpace(rawID)))
	hexStr := strings.ToUpper(hex.EncodeToString(hash[:]))

	// Formata como JUR-XXXX-XXXX-XXXX-XXXX para fácil identificação
	return fmt.Sprintf("JUR-%s-%s-%s-%s", hexStr[0:4], hexStr[4:8], hexStr[8:12], hexStr[12:16])
}

func obterUUIDWindows() string {
	// Tentativa 1: WMIC / PowerShell para obter UUID da máquina
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", "(Get-CimInstance -ClassName Win32_ComputerSystemProduct).UUID")
	saida, err := cmd.Output()
	if err == nil {
		uuid := strings.TrimSpace(string(saida))
		if len(uuid) > 10 && !strings.EqualFold(uuid, "FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF") {
			return uuid
		}
	}

	// Tentativa 2: Fallback para wmic clássico
	cmdWmic := exec.Command("wmic", "csproduct", "get", "uuid")
	saidaWmic, errWmic := cmdWmic.Output()
	if errWmic == nil {
		linhas := strings.Split(string(saidaWmic), "\n")
		for _, l := range linhas {
			l = strings.TrimSpace(l)
			if l != "" && !strings.EqualFold(l, "UUID") && len(l) > 10 {
				return l
			}
		}
	}

	return ""
}

func obterUUIDLinux() string {
	caminhos := []string{
		"/sys/class/dmi/id/product_uuid",
		"/etc/machine-id",
		"/var/lib/dbus/machine-id",
	}

	for _, cam := range caminhos {
		conteudo, err := os.ReadFile(cam)
		if err == nil {
			val := strings.TrimSpace(string(conteudo))
			if len(val) > 10 {
				return val
			}
		}
	}

	return ""
}

func obterFallbackHardware() string {
	var partes []string

	hostname, err := os.Hostname()
	if err == nil {
		partes = append(partes, hostname)
	}

	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			// Ignora interfaces de loopback ou virtuais sem MAC
			if iface.Flags&net.FlagLoopback == 0 && len(iface.HardwareAddr) > 0 {
				partes = append(partes, iface.HardwareAddr.String())
			}
		}
	}

	if len(partes) == 0 {
		return "DESKTOP-GENERICO-" + runtime.GOARCH
	}

	return strings.Join(partes, "|")
}
