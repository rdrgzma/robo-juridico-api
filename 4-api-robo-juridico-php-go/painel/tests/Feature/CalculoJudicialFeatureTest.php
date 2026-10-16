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

    public function test_compatibilidade_contrato_go_itens_liquidados_e_datas_br(): void
    {
        $payloadCalculo = [
            'numero_processo' => '5053283-30.2024.8.13.0079',
            'data_atualizacao' => '2026-10-14',
            'indexador_codigo' => 'TJMG',
        ];

        // Simula o formato nativo do Go (itens_liquidados, total_geral_devido e data DD/MM/AAAA)
        $resultadoGoNativo = [
            'sucesso' => true,
            'resumo' => [
                'total_geral_devido' => 12450.80,
                'total_principal_corrigido' => 10500.00,
                'total_juros_mora' => 1950.80,
            ],
            'itens_liquidados' => [
                [
                    'numero_item' => 1,
                    'descricao' => 'Condenação TJMG',
                    'data_vencimento' => '15/01/2024',
                    'valor_original' => 10000.00,
                    'fator_correcao' => 1.05,
                    'valor_corrigido' => 10500.00,
                    'valor_juros_mora' => 1950.80,
                    'valor_multa_moratoria' => 0.0,
                    'total_item' => 12450.80,
                ],
            ],
            'notas_metodologicas' => ['Atualizado pela Tabela da CGJ-MG'],
        ];

        $service = app(CalculoJudicialService::class);
        $salvo = $service->salvarCalculoComLead($payloadCalculo, $resultadoGoNativo, [
            'nome' => 'Dr. João Perito',
            'email' => 'joao@pericia.com.br',
            'whatsapp' => '31988887777',
            'oab' => 'OAB/MG 100.200',
        ]);

        $this->assertEquals(12450.80, (float) $salvo->total_geral);
        $this->assertEquals('novo', $salvo->status_lead);

        // Impressão do laudo não pode quebrar com data no formato DD/MM/AAAA
        $respLaudo = $this->get("/laudo/{$salvo->uuid}/imprimir");
        $respLaudo->assertStatus(200);
        $respLaudo->assertSee('12.450,80');
        $respLaudo->assertSee('15/01/2024');

        // Validação pública deve confirmar integridade
        $verificacao = $service->verificarAutenticidade($salvo->uuid);
        $this->assertTrue($verificacao['valido']);
    }

    public function test_gerador_qrcode_svg_nativo_e_mapeamento_tribunal(): void
    {
        $service = app(CalculoJudicialService::class);
        $dataUri = $service->gerarQrCodeDataUri('https://robojuridico.com.br/validar/teste-uuid');

        $this->assertStringStartsWith('data:image/svg+xml;base64,', $dataUri);
        $svgDecoded = base64_decode(substr($dataUri, strlen('data:image/svg+xml;base64,')));
        $this->assertStringContainsString('<svg', $svgDecoded);
        $this->assertStringContainsString('<rect', $svgDecoded);

        $this->assertEquals('TJMG', $service->mapearTribunalParaIndexador('TJMG'));
        $this->assertEquals('JF_GERAL', $service->mapearTribunalParaIndexador('TRF1'));
        $this->assertEquals('TST_TRABALHISTA', $service->mapearTribunalParaIndexador('TRT3'));
    }
}
