<?php

namespace Tests\Feature;

use App\Models\CalculoSalvo;
use App\Services\CalculoJudicialService;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Tests\TestCase;

class CalculoJudicialFeatureTest extends TestCase
{
    use RefreshDatabase;

    public function test_pagina_publica_calculadora_carrega_com_sucesso(): void
    {
        $response = $this->get('/');
        $response->assertStatus(200);
        $response->assertSee('Calculadora e Liquidação Judicial');
        $response->assertSee('TJSP');
        $response->assertSee('Importar Planilha');
    }

    public function test_emitir_laudo_salva_lead_e_gera_hash(): void
    {
        $payloadCalculo = [
            'numero_processo' => '0001234-56.2023.8.26.0100',
            'data_atualizacao' => '2026-10-14',
            'indexador_codigo' => 'TJSP',
            'parcelas' => [
                [
                    'descricao' => 'Parcela 01',
                    'tipo' => 'DEBITO',
                    'data_vencimento' => '2024-01-15',
                    'valor_original' => 5000.00,
                ],
            ],
        ];

        $resultadoSimulado = [
            'sucesso' => true,
            'resumo' => [
                'total_geral' => 5890.45,
                'total_principal_corrigido' => 5400.00,
                'total_juros' => 490.45,
            ],
            'itens' => [
                [
                    'numero_item' => 1,
                    'descricao' => 'Parcela 01',
                    'data_vencimento' => '2024-01-15',
                    'valor_original' => 5000.00,
                    'fator_correcao' => 1.08,
                    'valor_corrigido' => 5400.00,
                    'valor_juros' => 490.45,
                    'valor_multa_moratoria' => 0.0,
                    'subtotal' => 5890.45,
                ],
            ],
            'notas_legais' => ['Atualização com Tabela TJSP'],
        ];

        $response = $this->postJson('/calculadora/emitir-laudo', [
            'nome' => 'Dra. Maria Advogada',
            'email' => 'maria@advocacia.com.br',
            'whatsapp' => '(11) 99999-8888',
            'oab' => 'OAB/SP 456.789',
            'dados_calculo' => $payloadCalculo,
            'resultado' => $resultadoSimulado,
        ]);

        $response->assertStatus(200);
        $response->assertJsonStructure(['ok', 'uuid', 'hash', 'url_laudo', 'url_validacao']);

        $uuid = $response->json('uuid');
        $this->assertDatabaseHas('calculos_salvos', [
            'uuid' => $uuid,
            'lead_nome' => 'Dra. Maria Advogada',
            'lead_email' => 'maria@advocacia.com.br',
            'total_geral' => 5890.45,
        ]);

        // Testa visualização do Laudo para impressão
        $responseLaudo = $this->get("/laudo/{$uuid}/imprimir");
        $responseLaudo->assertStatus(200);
        $responseLaudo->assertSee('Dra. Maria Advogada');
        $responseLaudo->assertSee('0001234-56.2023.8.26.0100');
        $responseLaudo->assertSee('5.890,45');

        // Testa validação pública de autenticidade
        $responseValidacao = $this->get("/validar/{$uuid}");
        $responseValidacao->assertStatus(200);
        $responseValidacao->assertSee('Cálculo Autêntico e Íntegro');
    }

    public function test_validacao_rejeita_uuid_inexistente(): void
    {
        $response = $this->get('/validar/00000000-0000-0000-0000-000000000000');
        $response->assertStatus(200);
        $response->assertSee('Autenticidade Não Confirmada');
    }
}
