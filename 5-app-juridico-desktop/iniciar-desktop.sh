#!/usr/bin/env bash
set -e

echo "=================================================================="
echo "   INICIANDO ROBO JURIDICO DESKTOP ON-PREMISE (LINUX)"
echo "=================================================================="

if [ -f "./robo-juridico-desktop-linux" ]; then
    chmod +x ./robo-juridico-desktop-linux
    ./robo-juridico-desktop-linux
else
    go run ./cmd/app
fi
