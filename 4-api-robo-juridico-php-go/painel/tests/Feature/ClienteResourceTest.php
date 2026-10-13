<?php

namespace Tests\Feature;

use App\Models\Chave;
use App\Models\Cliente;
use App\Models\Uso;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class ClienteResourceTest extends TestCase
{
    use RefreshDatabase;

    public function test_can_create_and_query_cliente(): void
    {
        $cliente = Cliente::create([
            'nome' => 'Escritório Silva & Advogados',
            'email' => 'contato@silva.adv.br',
            'empresa' => 'Silva Sociedade de Advogados',
            'plano' => 'enterprise',
            'valor_mensal_centavos' => 59900,
            'limite_mensal' => 50000,
            'limite_por_minuto' => 120,
            'ativo' => true,
            'observacoes' => 'Cliente prioritário VIP',
            'criado_em' => now()->utc()->toIso8601String(),
        ]);

        $this->assertDatabaseHas('clientes', [
            'email' => 'contato@silva.adv.br',
            'empresa' => 'Silva Sociedade de Advogados',
            'plano' => 'enterprise',
        ]);

        $this->assertEquals(50000, $cliente->saldoRestante());
        $this->assertEquals(0.0, $cliente->porcentagemConsumo());
    }

    public function test_cliente_relationships(): void
    {
        $cliente = Cliente::create([
            'nome' => 'Juris Tech',
            'email' => 'juristech@exemplo.com',
            'plano' => 'pro',
            'limite_mensal' => 10000,
            'ativo' => true,
            'criado_em' => now()->utc()->toIso8601String(),
        ]);

        $chave = Chave::create([
            'cliente_id' => $cliente->id,
            'nome' => 'Chave Integração',
            'prefixo' => 'rj_live_abc12345...',
            'hash' => hash('sha256', 'rj_live_secret_key_mock'),
            'criada_em' => now()->utc()->toIso8601String(),
        ]);

        $uso = Uso::create([
            'cliente_id' => $cliente->id,
            'chave_id' => $chave->id,
            'momento' => now()->utc()->toIso8601String(),
            'metodo' => 'POST',
            'rota' => '/api/v1/processos/consultar',
            'caminho' => '/api/v1/processos/consultar',
            'status' => 200,
            'ms' => 120,
            'ip' => '127.0.0.1',
            'cobravel' => true,
            'cobrado' => true,
        ]);

        $this->assertCount(1, $cliente->chaves);
        $this->assertCount(1, $cliente->usos);
        $this->assertEquals($cliente->id, $uso->cliente->id);
        $this->assertEquals($chave->id, $uso->chave->id);
    }
}
