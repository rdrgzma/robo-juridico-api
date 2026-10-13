<?php

namespace Tests\Unit;

use App\Models\Chave;
use App\Models\Cliente;
use App\Models\Uso;
use Carbon\Carbon;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class ConsumoCalculoTest extends TestCase
{
    use RefreshDatabase;

    public function test_consumo_mes_atual_calculates_correctly_in_brasilia_timezone(): void
    {
        $cliente = Cliente::create([
            'nome' => 'Advocacia Pinheiro',
            'email' => 'contato@pinheiro.adv.br',
            'plano' => 'pro',
            'limite_mensal' => 100,
            'ativo' => true,
            'criado_em' => now()->utc()->toIso8601String(),
        ]);

        $chave = Chave::create([
            'cliente_id' => $cliente->id,
            'nome' => 'Chave 1',
            'prefixo' => 'rj_live_test...',
            'hash' => 'hash_teste',
            'criada_em' => now()->utc()->toIso8601String(),
        ]);

        $agora = Carbon::now('America/Sao_Paulo');

        // 1. Consulta no mês atual (cobrável, status 200) -> DEVE contar
        Uso::create([
            'cliente_id' => $cliente->id,
            'chave_id' => $chave->id,
            'momento' => $agora->copy()->utc()->toIso8601String(),
            'metodo' => 'POST',
            'rota' => '/api/v1/processos/consultar',
            'caminho' => '/api/v1/processos/consultar',
            'status' => 200,
            'ms' => 80,
            'cobravel' => true,
            'cobrado' => true,
        ]);

        // 2. Outra consulta no mês atual (cobrável, status 200) -> DEVE contar
        Uso::create([
            'cliente_id' => $cliente->id,
            'chave_id' => $chave->id,
            'momento' => $agora->copy()->subDays(2)->utc()->toIso8601String(),
            'metodo' => 'POST',
            'rota' => '/api/v1/processos/consultar',
            'caminho' => '/api/v1/processos/consultar',
            'status' => 200,
            'ms' => 95,
            'cobravel' => true,
            'cobrado' => true,
        ]);

        // 3. Consulta no mês PASSADO -> NÃO DEVE contar no consumo do mês atual
        Uso::create([
            'cliente_id' => $cliente->id,
            'chave_id' => $chave->id,
            'momento' => $agora->copy()->subMonths(2)->utc()->toIso8601String(),
            'metodo' => 'POST',
            'rota' => '/api/v1/processos/consultar',
            'caminho' => '/api/v1/processos/consultar',
            'status' => 200,
            'ms' => 90,
            'cobravel' => true,
            'cobrado' => true,
        ]);

        // 4. Consulta não cobrável (ex: erro 400 ou status 402) -> NÃO DEVE contar
        Uso::create([
            'cliente_id' => $cliente->id,
            'chave_id' => $chave->id,
            'momento' => $agora->copy()->utc()->toIso8601String(),
            'metodo' => 'POST',
            'rota' => '/api/v1/processos/consultar',
            'caminho' => '/api/v1/processos/consultar',
            'status' => 402,
            'ms' => 5,
            'cobravel' => false,
            'cobrado' => false,
        ]);

        $this->assertEquals(2, $cliente->consumoMesAtual());
        $this->assertEquals(98, $cliente->saldoRestante());
        $this->assertEquals(2.0, $cliente->porcentagemConsumo());
    }

    public function test_saldo_restante_never_negative(): void
    {
        $cliente = Cliente::create([
            'nome' => 'Cliente Limite Baixo',
            'email' => 'baixo@exemplo.com',
            'plano' => 'starter',
            'limite_mensal' => 1,
            'ativo' => true,
            'criado_em' => now()->utc()->toIso8601String(),
        ]);

        $chave = Chave::create([
            'cliente_id' => $cliente->id,
            'nome' => 'Chave 1',
            'prefixo' => 'rj_live_test...',
            'hash' => 'hash_teste2',
            'criada_em' => now()->utc()->toIso8601String(),
        ]);

        // 2 consultas cobráveis
        for ($i = 0; $i < 2; $i++) {
            Uso::create([
                'cliente_id' => $cliente->id,
                'chave_id' => $chave->id,
                'momento' => now()->utc()->toIso8601String(),
                'metodo' => 'POST',
                'rota' => '/api/v1/processos/consultar',
                'caminho' => '/api/v1/processos/consultar',
                'status' => 200,
                'ms' => 50,
                'cobravel' => true,
                'cobrado' => true,
            ]);
        }

        $this->assertEquals(2, $cliente->consumoMesAtual());
        $this->assertEquals(0, $cliente->saldoRestante());
        $this->assertEquals(100.0, $cliente->porcentagemConsumo());
    }
}
