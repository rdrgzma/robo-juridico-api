#!/bin/sh
set -e

# Garante existência do arquivo .env
if [ ! -f /app/.env ]; then
    cp /app/.env.example /app/.env
fi

# Garante diretório de dados para persistência do SQLite
mkdir -p /app/dados

# Executa as migrations do Laravel se necessário
php artisan migrate --force

# Executa o seeder para garantir o admin inicial
php artisan db:seed --force

echo "======================================================="
echo "Painel Administrativo Robô Jurídico (Filament + Laravel)"
echo "Iniciando na porta 8080 (http://0.0.0.0:8080)..."
echo "======================================================="

exec php artisan serve --host=0.0.0.0 --port=8080
