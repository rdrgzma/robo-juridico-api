<?php

namespace Tests\Feature;

use App\Models\Cliente;
use App\Services\ChaveService;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\Auth;
use Tests\TestCase;

class ClienteChaveSelfServiceTest extends TestCase
{
    use RefreshDatabase;

    public function test_cliente_can_generate_and_revoke_own_key_self_service(): void
    {
        $cliente = Cliente::create([
            'nome' => 'Tech Law Solutions',
            'email' => 'dev@techlaw.com.br',
            'plano' => 'enterprise',
            'ativo' => true,
            'criado_em' => now()->utc()->toIso8601String(),
            'atualizado_em' => now()->utc()->toIso8601String(),
        ]);

        Auth::guard('cliente')->login($cliente);

        $service = app(ChaveService::class);
        $resultado = $service->gerarParaCliente($cliente, 'Chave de Produção 2026');

        $token = $resultado['token'];
        $chave = $resultado['chave'];

        $this->assertStringStartsWith('rj_live_', $token);
        $this->assertEquals($cliente->id, $chave->cliente_id);
        $this->assertEquals('Chave de Produção 2026', $chave->nome);
        $this->assertTrue($chave->isAtiva());

        // Revogação
        $chave->revogar();
        $this->assertFalse($chave->fresh()->isAtiva());
        $this->assertNotNull($chave->fresh()->revogada_em);
    }
}
