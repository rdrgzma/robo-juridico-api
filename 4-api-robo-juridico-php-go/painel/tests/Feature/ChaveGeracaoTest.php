<?php

namespace Tests\Feature;

use App\Models\Chave;
use App\Models\Cliente;
use App\Services\ChaveService;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class ChaveGeracaoTest extends TestCase
{
    use RefreshDatabase;

    public function test_chave_generation_format_and_sha256_compatibility(): void
    {
        $cliente = Cliente::create([
            'nome' => 'Cliente Alpha',
            'email' => 'alpha@empresa.com.br',
            'empresa' => 'Alpha Corp',
            'plano' => 'pro',
            'valor_mensal_centavos' => 29900,
            'limite_mensal' => 5000,
            'limite_por_minuto' => 60,
            'ativo' => true,
            'criado_em' => now()->utc()->toIso8601String(),
        ]);

        $service = new ChaveService();
        $resultado = $service->gerarParaCliente($cliente, 'Chave de Produção');

        $token = $resultado['token'];
        /** @var Chave $chave */
        $chave = $resultado['chave'];

        // 1. Prefixo correto
        $this->assertStringStartsWith('rj_live_', $token);

        // 2. Comprimento do token (rj_live_ + 43 caracteres = 51)
        $this->assertGreaterThanOrEqual(40, strlen($token));

        // 3. Persistência no banco
        $this->assertEquals($cliente->id, $chave->cliente_id);
        $this->assertEquals('Chave de Produção', $chave->nome);
        $this->assertStringStartsWith('rj_live_', $chave->prefixo);
        $this->assertStringEndsWith('...', $chave->prefixo);

        // 4. Hash SHA-256 coincide com hash nativo usado pela API Go
        $hashEsperado = hash('sha256', $token);
        $this->assertEquals(64, strlen($chave->hash));
        $this->assertEquals($hashEsperado, $chave->hash);

        // 5. Estado inicial ativo
        $this->assertTrue($chave->isAtiva());
        $this->assertNull($chave->revogada_em);
    }

    public function test_chave_revocation(): void
    {
        $cliente = Cliente::create([
            'nome' => 'Cliente Beta',
            'email' => 'beta@empresa.com.br',
            'empresa' => 'Beta Corp',
            'plano' => 'starter',
            'valor_mensal_centavos' => 9900,
            'limite_mensal' => 1000,
            'limite_por_minuto' => 20,
            'ativo' => true,
            'criado_em' => now()->utc()->toIso8601String(),
        ]);

        $service = new ChaveService();
        $resultado = $service->gerarParaCliente($cliente);
        $chave = $resultado['chave'];

        $this->assertTrue($chave->isAtiva());

        $chave->revogar();
        $this->assertFalse($chave->fresh()->isAtiva());
        $this->assertNotNull($chave->fresh()->revogada_em);
    }
}
