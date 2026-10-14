@echo off
chcp 65001 >nul
echo =======================================================
echo Executando suite de testes automatizados da API em Go...
echo =======================================================
cd /d "%~dp0\4-api-robo-juridico-php-go\api-juridica-go"
go test -v ./...
if %errorlevel% neq 0 (
    echo.
    echo [FALHA] Alguns testes falharam! Verifique os erros acima.
    exit /b %errorlevel%
)
echo.
echo [SUCESSO] Todos os testes unitários e de integração passaram com sucesso!
