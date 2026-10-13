# ==============================================================================
# Dockerfile: Robô Jurídico API (consulta-processos)
# Multi-stage / Production-ready Python 3.12-slim
# ==============================================================================
FROM python:3.12-slim

LABEL maintainer="Robô Jurídico" \
      description="API de Consulta Unificada de Processos Judiciais nas Fontes Oficiais do CNJ (DataJud + DJEN)" \
      version="0.1.0"

# Otimizações de runtime do Python
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1 \
    API_BANCO=/app/dados/api.db

# Instalação de dependências de sistema para compilação (curl-cffi) e healthcheck
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libffi-dev \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Usuário de sistema não-root por segurança
RUN groupadd -g 1000 appgroup && \
    useradd -u 1000 -g appgroup -s /bin/bash -m appuser

# Copia metadados e arquivos do pacote a partir da pasta de código
COPY 3-api-robo-juridico/api-robo-juridico/pyproject.toml \
     3-api-robo-juridico/api-robo-juridico/README.md \
     3-api-robo-juridico/api-robo-juridico/LICENSE ./
COPY 3-api-robo-juridico/api-robo-juridico/consulta_processos ./consulta_processos

# Instala a aplicação com extensões de API REST (FastAPI) e Navegador (curl-cffi)
RUN pip install --upgrade pip && \
    pip install -e ".[api,navegador]"

# Cria diretório para persistência do banco SQLite
RUN mkdir -p /app/dados && \
    chown -R appuser:appgroup /app

USER appuser

# Porta padrão de escuta da API REST
EXPOSE 8000

# Verificação de vida pelo endpoint nativo /saude
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8000/saude || exit 1

# Comando de inicialização
CMD ["consulta-processos", "api", "--host", "0.0.0.0", "--porta", "8000"]
