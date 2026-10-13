<?php

namespace Tests\Feature;

use App\Filament\Portal\Resources\MeuUsoResource;
use App\Filament\Portal\Resources\MinhaChaveResource;
use App\Models\Chave;
use App\Models\Cliente;
use App\Models\Uso;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\Auth;
use Tests\TestCase;

class ClientePortalIsolationTest extends TestCase
{
    use RefreshDatabase;

    public function test_portal_strictly_isolates_api_keys_between_clients(): void
    {
        $clienteA = Cliente::create([
            'nome' => 'Empresa Alfa',
            'email' => 'alfa@empresa.com',
            'plano' => 'pro',
            'ativo' => true,
            'criado_em' => now()->utc()->toIso8601String(),
            'atualizado_em' => now()->utc()->toIso8601String(),
        ]);

        $clienteB = Cliente::create([
            'nome' => 'Empresa Beta',
            'email' => 'beta@empresa.com',
            'plano' => 'basico',
            'ativo' => true,
            'criado_em' => now()->utc()->toIso8601String(),
            'atualizado_em' => now()->utc()->toIso8601String(),
        ]);

        $chaveA = Chave::create([
            'cliente_id' => $clienteA->id,
            'nome' => 'Chave Secreta Alfa',
            'prefixo' => 'rj_live_alfa...',
            'hash' => 'hash_alfa',
            'criada_em' => now()->utc()->toIso8601String(),
        ]);

        $chaveB = Chave::create([
            'cliente_id' => $clienteB->id,
            'nome' => 'Chave Secreta Beta',
            'prefixo' => 'rj_live_beta...',
            'hash' => 'hash_beta',
            'criada_em' => now()->utc()->toIso8601String(),
        ]);

        // Autentica como Cliente A
        Auth::guard('cliente')->login($clienteA);

        $chavesVisiveis = MinhaChaveResource::getEloquentQuery()->pluck('id')->all();

        $this->assertContains($chaveA->id, $chavesVisiveis);
        $this->assertNotContains($chaveB->id, $chavesVisiveis, 'Cliente A não deve enxergar a chave do Cliente B');
    }

    public function test_portal_strictly_isolates_usage_logs_between_clients(): void
    {
        $clienteA = Cliente::create([
            'nome' => 'Empresa Alfa',
            'email' => 'alfa@empresa.com',
            'plano' => 'pro',
            'ativo' => true,
            'criado_em' => now()->utc()->toIso8601String(),
            'atualizado_em' => now()->utc()->toIso8601String(),
        ]);

        $clienteB = Cliente::create([
            'nome' => 'Empresa Beta',
            'email' => 'beta@empresa.com',
            'plano' => 'basico',
            'ativo' => true,
            'criado_em' => now()->utc()->toIso8601String(),
            'atualizado_em' => now()->utc()->toIso8601String(),
        ]);

        $usoA = Uso::create([
            'cliente_id' => $clienteA->id,
            'momento' => now()->utc()->toIso8601String(),
            'metodo' => 'POST',
            'rota' => '/api/v1/processos/consultar',
            'status' => 200,
            'ms' => 45,
            'cobravel' => true,
        ]);

        $usoB = Uso::create([
            'cliente_id' => $clienteB->id,
            'momento' => now()->utc()->toIso8601String(),
            'metodo' => 'POST',
            'rota' => '/api/v1/processos/consultar',
            'status' => 200,
            'ms' => 60,
            'cobravel' => true,
        ]);

        // Autentica como Cliente A
        Auth::guard('cliente')->login($clienteA);

        $usosVisiveis = MeuUsoResource::getEloquentQuery()->pluck('id')->all();

        $this->assertContains($usoA->id, $usosVisiveis);
        $this->assertNotContains($usoB->id, $usosVisiveis, 'Cliente A não deve enxergar histórico de requisições do Cliente B');
    }
}
