@echo off
title Robo Juridico Desktop On-Premise
cls
echo ==================================================================
echo    INICIANDO ROBO JURIDICO DESKTOP ON-PREMISE (WINDOWS)
echo ==================================================================
echo.

if exist "robo-juridico-desktop.exe" (
    echo Iniciando executavel compilado...
    start "" "robo-juridico-desktop.exe"
) else (
    echo Compilando e executando via Go...
    go run ./cmd/app
)
