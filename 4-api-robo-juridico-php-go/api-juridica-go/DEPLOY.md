# Guia Rápido de Deploy da API em Go (`api-juridica-go`)

Este documento descreve como realizar o deploy da API em Go nos ambientes de desenvolvimento, homologação e produção. Para o guia arquitetural completo, consulte [07-GUIA-DE-DEPLOY-DA-API.md](../../docs/07-GUIA-DE-DEPLOY-DA-API.md).

---

## 🚀 Opção 1: Deploy com Docker Compose (Recomendado)

### 1. Preparar o arquivo de variáveis
Crie um arquivo `.env` baseado no modelo abaixo:

```env
HOST=0.0.0.0
PORT=8000
API_BANCO=/app/dados/api.db
API_LIMITE_POR_MINUTO=60
DATAJUD_TIMEOUT=90
COMUNICA_TIMEOUT=12
API_CHAVES=admin:minha-chave-secreta-longa-com-mais-de-32-chars
```

### 2. Construir e Iniciar os Containers
```bash
docker compose up -d --build
```

### 3. Verificar se está funcionando
```bash
curl http://localhost:8000/saude
```
Retorno:
```json
{"ok": true, "versao": "1.0.0"}
```

Para abrir a documentação Swagger interativa:
Acesse no navegador: `http://localhost:8000/docs`

---

## ⚡ Opção 2: Deploy Nativo no Linux (Systemd)

### 1. Compilar o Binário Estático
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -trimpath \
  -ldflags="-s -w -X main.version=1.0.0" \
  -o api-juridica ./cmd/api
```

### 2. Instalar no Servidor
```bash
sudo useradd -r -s /bin/false -d /opt/robo-juridico robojuridico
sudo mkdir -p /opt/robo-juridico/bin /opt/robo-juridico/dados
sudo mv api-juridica /opt/robo-juridico/bin/api
sudo chmod +x /opt/robo-juridico/bin/api
sudo chown -R robojuridico:robojuridico /opt/robo-juridico
```

### 3. Configurar Serviço no Systemd
Crie `/etc/systemd/system/robo-juridico.service`:
```ini
[Unit]
Description=Robô Jurídico API
After=network.target

[Service]
Type=simple
User=robojuridico
Group=robojuridico
WorkingDirectory=/opt/robo-juridico
ExecStart=/opt/robo-juridico/bin/api
Restart=always
RestartSec=5s
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

Ative e inicie:
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now robo-juridico
```

---

## 🔒 Proxy Reverso com Caddy (HTTPS Automático)

No arquivo `/etc/caddy/Caddyfile`:
```caddy
api.seudominio.com.br {
    reverse_proxy 127.0.0.1:8000
}
```
Recarregue o Caddy:
```bash
sudo systemctl reload caddy
```
O Caddy cuidará automaticamente do certificado SSL (Let's Encrypt).
