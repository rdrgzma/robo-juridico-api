<?php

namespace Database\Seeders;

use App\Models\Cliente;
use App\Services\ChaveService;
use Carbon\Carbon;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\Hash;

class ClienteSeeder extends Seeder
{
    public function run(): void
    {
        $agora = Carbon::now()->utc()->toIso8601String();

        $cliente = Cliente::firstOrCreate(
            ['email' => 'cliente@empresa.com.br'],
            [
                'nome' => 'Carlos Silva',
                'empresa' => 'Silva & Associados Advocacia',
                'senha_hash' => Hash::make('cliente123'),
                'plano' => 'pro',
                'valor_mensal_centavos' => 29900,
                'limite_mensal' => 5000,
                'limite_por_minuto' => 60,
                'ativo' => true,
                'observacoes' => 'Cliente de demonstração do portal',
                'criado_em' => $agora,
                'atualizado_em' => $agora,
            ]
        );

        // Se ainda não tiver chaves, gera uma chave inicial
        if ($cliente->chaves()->count() === 0) {
            $chaveService = app(ChaveService::class);
            $chaveService->gerarParaCliente($cliente, 'Chave de Testes e Integração');
        }
    }
}
