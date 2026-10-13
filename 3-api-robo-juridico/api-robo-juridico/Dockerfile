# ==============================================================================
# Dockerfile: Robô Jurídico API (consulta-processos)
# Multi-stage / Production-ready Python 3.12-slim
# ==============================================================================
FROM python:3.12-slim

LABEL maintainer="Robô Jurídico" \
      description="API de Consulta Unificada de Processos Judiciais nas Fontes Oficiais do CNJ (DataJud + DJEN)" \
      version="0.1.0"

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1 \
    API_BANCO=/app/dados/api.db

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libffi-dev \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN groupadd -g 1000 appgroup && \
    useradd -u 1000 -g appgroup -s /bin/bash -m appuser

COPY pyproject.toml README.md LICENSE ./
COPY consulta_processos ./consulta_processos

RUN pip install --upgrade pip && \
    pip install -e ".[api,navegador]"

RUN mkdir -p /app/dados && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8000/saude || exit 1

CMD ["consulta-processos", "api", "--host", "0.0.0.0", "--porta", "8000"]
