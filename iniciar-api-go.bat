@echo off
chcp 65001 >nul
echo =======================================================
echo Iniciando Robô Jurídico API (Go) localmente...
echo =======================================================
cd /d "%~dp0\4-api-robo-juridico-php-go\api-juridica-go"
go run cmd/api/main.go
