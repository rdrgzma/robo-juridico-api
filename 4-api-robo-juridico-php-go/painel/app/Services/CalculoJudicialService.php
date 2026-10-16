<?php

namespace App\Services;

use App\Models\CalculoSalvo;
use chillerlan\QRCode\QRCode;
use chillerlan\QRCode\QROptions;
use Illuminate\Http\UploadedFile;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Str;

class CalculoJudicialService
{
    protected string $baseUrl;
    protected string $apiKey;

    public function __construct()
    {
        $this->baseUrl = rtrim(env('GO_API_URL', 'http://127.0.0.1:8000'), '/');
        $this->apiKey = env('GO_API_KEY', 'teste-chave-secreta-longa-com-mais-de-32-caracteres');
    }

    /**
     * Executa a liquidação judicial completa enviando a requisição para a API em Go
     */
    public function calcular(array $dados): array
    {
        try {
            $response = Http::timeout(10)
                ->withHeaders([
                    'X-API-Key' => $this->apiKey,
                    'Accept' => 'application/json',
                    'Content-Type' => 'application/json',
                ])
                ->post("{$this->baseUrl}/v1/calculos/judicial", $dados);

            if ($response->successful()) {
                return $response->json();
            }

            $erro = $response->json('erro.mensagem') ?? $response->json('detalhe') ?? $response->json('mensagem') ?? $response->body();
            return [
                'sucesso' => false,
                'mensagem' => "Falha no cálculo ({$response->status()}): {$erro}",
            ];
        } catch (\Throwable $e) {
            Log::error("[CALCULO_JUDICIAL] Erro ao conectar na API Go: " . $e->getMessage());
            return [
                'sucesso' => false,
                'mensagem' => "Não foi possível conectar ao motor de cálculo em Go ({$this->baseUrl}): " . $e->getMessage(),
            ];
        }
    }

    /**
     * Importa planilha (.xlsx, .csv, .tsv) ou texto colado via Go Engine
     */
    public function importarPlanilha($arquivoOuConteudo, ?string $nomeArquivo = null): array
    {
        try {
            $req = Http::timeout(15)->withHeaders([
                'X-API-Key' => $this->apiKey,
                'Accept' => 'application/json',
            ]);

            if ($arquivoOuConteudo instanceof UploadedFile) {
                $response = $req->attach(
                    'planilha',
                    file_get_contents($arquivoOuConteudo->getRealPath()),
                    $arquivoOuConteudo->getClientOriginalName()
                )->post("{$this->baseUrl}/v1/calculos/importar-planilha");
            } elseif (is_string($arquivoOuConteudo) && @file_exists($arquivoOuConteudo)) {
                $nome = $nomeArquivo ?: basename($arquivoOuConteudo);
                $response = $req->attach(
                    'planilha',
                    file_get_contents($arquivoOuConteudo),
                    $nome
                )->post("{$this->baseUrl}/v1/calculos/importar-planilha");
            } else {
                // Conteúdo textual ou CSV em string pura
                $nome = $nomeArquivo ?: 'dados_colados.csv';
                $response = $req->withHeaders([
                    'X-Filename' => $nome,
                    'Content-Type' => 'text/plain; charset=utf-8',
                ])->withBody((string)$arquivoOuConteudo, 'text/plain')
                  ->post("{$this->baseUrl}/v1/calculos/importar-planilha");
            }

            if ($response->successful()) {
                return $response->json();
            }

            $erro = $response->json('erro.mensagem') ?? $response->json('detalhe') ?? $response->json('mensagem') ?? $response->body();
            return [
                'ok' => false,
                'mensagem' => "Erro ao importar planilha ({$response->status()}): {$erro}",
            ];
        } catch (\Throwable $e) {
            Log::error("[CALCULO_JUDICIAL] Erro ao importar planilha: " . $e->getMessage());
            return [
                'ok' => false,
                'mensagem' => "Falha de comunicação ao importar: " . $e->getMessage(),
            ];
        }
    }

    /**
     * Exporta a memória de cálculo liquidada para planilha Excel (.xlsx) ou CSV (.csv)
     */
    public function exportarPlanilha(array $dadosOuResultado, string $formato = 'xlsx'): ?string
    {
        try {
            $formato = strtolower(trim($formato)) === 'csv' ? 'csv' : 'xlsx';
            $response = Http::timeout(15)
                ->withHeaders([
                    'X-API-Key' => $this->apiKey,
                    'Content-Type' => 'application/json',
                ])
                ->post("{$this->baseUrl}/v1/calculos/exportar-planilha?formato={$formato}", $dadosOuResultado);

            if ($response->successful()) {
                return $response->body();
            }
        } catch (\Throwable $e) {
            Log::warning("[CALCULO_JUDICIAL] Falha ao exportar planilha: " . $e->getMessage());
        }

        return null;
    }

    /**
     * Salva o cálculo realizado vinculando ao Lead/Cliente e gerando hash criptográfico de autenticidade
     */
    public function salvarCalculoComLead(
        array $dadosCalculo,
        array $resultado,
        array $leadData = [],
        ?string $ipOrigem = null,
        ?int $clienteId = null
    ): CalculoSalvo {
        $uuid = Str::uuid()->toString();

        if ($clienteId === null && Auth::guard('cliente')->check()) {
            $clienteId = (int) Auth::guard('cliente')->id();
        }

        $totalGeral = (float) ($resultado['resumo']['total_geral']
            ?? $resultado['resumo']['total_geral_devido']
            ?? 0);
        $itensLista = $resultado['itens'] ?? $resultado['itens_liquidados'] ?? [];

        // Garante que o JSON salvo contenha ambas as chaves de compatibilidade
        $resultado['resumo']['total_geral'] = $totalGeral;
        $resultado['resumo']['total_geral_devido'] = $totalGeral;
        $resultado['itens'] = $itensLista;
        $resultado['itens_liquidados'] = $itensLista;

        $hashBase = json_encode([
            'uuid' => $uuid,
            'processo' => $dadosCalculo['numero_processo'] ?? '',
            'data_atualizacao' => $dadosCalculo['data_atualizacao'] ?? '',
            'total_geral' => $totalGeral,
            'itens_count' => count($itensLista),
        ]);
        $hashValidacao = hash('sha256', $hashBase . 'ROBO_JURIDICO_SALT_2026');

        return CalculoSalvo::create([
            'cliente_id' => $clienteId,
            'uuid' => $uuid,
            'hash_validacao' => $hashValidacao,
            'numero_processo' => $dadosCalculo['numero_processo'] ?? null,
            'data_atualizacao' => $dadosCalculo['data_atualizacao'] ?? now()->format('Y-m-d'),
            'indexador_codigo' => $dadosCalculo['indexador_codigo'] ?? 'TJSP',
            'total_geral' => $totalGeral,
            'payload_requisicao' => $dadosCalculo,
            'resultado_json' => $resultado,
            'lead_nome' => $leadData['nome'] ?? null,
            'lead_email' => $leadData['email'] ?? null,
            'lead_whatsapp' => $leadData['whatsapp'] ?? null,
            'lead_oab' => $leadData['oab'] ?? null,
            'status_lead' => $leadData['status_lead'] ?? 'novo',
            'ip_origem' => $ipOrigem ?? request()->ip(),
        ]);
    }

    /**
     * Recalcula um cálculo salvo atualizando a data-base para hoje (1-Click Update)
     */
    public function atualizarCalculoParaHoje(CalculoSalvo $calculo): ?CalculoSalvo
    {
        $payload = $calculo->payload_requisicao ?? [];
        $payload['data_atualizacao'] = now()->format('Y-m-d');

        // Se o payload veio do formato de formulário do Filament, normaliza para o formato da API Go
        if (!isset($payload['juros_mora']) && isset($payload['juros_tipo'])) {
            $payload = [
                'numero_processo' => $payload['numero_processo'] ?? $calculo->numero_processo,
                'data_atualizacao' => now()->format('Y-m-d'),
                'indexador_codigo' => $payload['indexador_codigo'] ?? $calculo->indexador_codigo,
                'aplicar_deflacao' => (bool)($payload['aplicar_deflacao'] ?? false),
                'juros_mora' => [
                    'tipo' => $payload['juros_tipo'] ?? 'CONVENCIONAL_MENSAL',
                    'percentual_mensal' => (float)($payload['juros_percentual'] ?? 1.0),
                    'modo' => $payload['juros_modo'] ?? 'SIMPLES',
                    'aplicar_prorata_die' => (bool)($payload['juros_prorata'] ?? true),
                ],
                'fase_execucao' => [
                    'aplicar_art_523_multa' => (bool)($payload['aplicar_art_523_multa'] ?? false),
                    'aplicar_art_523_honorarios' => (bool)($payload['aplicar_art_523_honorarios'] ?? false),
                    'multa_moratoria_percentual' => (float)($payload['multa_moratoria_percentual'] ?? 0.0),
                    'honorarios_sucumbenciais_percentual' => (float)($payload['honorarios_sucumbenciais_percentual'] ?? 0.0),
                ],
                'parcelas' => $payload['parcelas'] ?? [],
            ];
        }

        $novoResultado = $this->calcular($payload);
        if (empty($novoResultado['sucesso'])) {
            return null;
        }

        return $this->salvarCalculoComLead(
            $payload,
            $novoResultado,
            [
                'nome' => $calculo->lead_nome,
                'email' => $calculo->lead_email,
                'whatsapp' => $calculo->lead_whatsapp,
                'oab' => $calculo->lead_oab,
                'status_lead' => $calculo->status_lead ?? 'novo',
            ],
            $calculo->ip_origem,
            $calculo->cliente_id
        );
    }

    /**
     * Busca um cálculo salvo pelo seu identificador público UUID
     */
    public function buscarPorUuid(string $uuid): ?CalculoSalvo
    {
        return CalculoSalvo::where('uuid', $uuid)->first();
    }

    /**
     * Valida se o laudo permanece íntegro e autêntico comparando o hash de verificação
     */
    public function verificarAutenticidade(string $uuid): array
    {
        $calculo = $this->buscarPorUuid($uuid);
        if (!$calculo) {
            return [
                'valido' => false,
                'motivo' => 'Cálculo não localizado na base pública de certificação.',
            ];
        }

        $dadosCalculo = $calculo->payload_requisicao ?? [];
        $resultado = $calculo->resultado_json ?? [];

        $totalGeral = (float) ($resultado['resumo']['total_geral']
            ?? $resultado['resumo']['total_geral_devido']
            ?? 0);
        $itensLista = $resultado['itens'] ?? $resultado['itens_liquidados'] ?? [];

        $hashBase = json_encode([
            'uuid' => $calculo->uuid,
            'processo' => $dadosCalculo['numero_processo'] ?? '',
            'data_atualizacao' => $dadosCalculo['data_atualizacao'] ?? '',
            'total_geral' => $totalGeral,
            'itens_count' => count($itensLista),
        ]);
        $hashEsperado = hash('sha256', $hashBase . 'ROBO_JURIDICO_SALT_2026');

        if ($hashEsperado !== $calculo->hash_validacao) {
            return [
                'valido' => false,
                'motivo' => 'Inconsistência de integridade: o hash do registro foi violado ou adulterado.',
                'calculo' => $calculo,
            ];
        }

        return [
            'valido' => true,
            'calculo' => $calculo,
            'data_emissao' => $calculo->created_at->format('d/m/Y H:i:s'),
            'hash_validacao' => $calculo->hash_validacao,
        ];
    }

    /**
     * Gera QR Code em formato Data URI (SVG vetorial real) para impressão no Laudo Pericial
     */
    public function gerarQrCodeDataUri(string $conteudo): string
    {
        try {
            if (class_exists(QRCode::class)) {
                $options = new QROptions([
                    'outputType' => QRCode::OUTPUT_MARKUP_SVG,
                    'eccLevel' => QRCode::ECC_L,
                    'scale' => 4,
                    'imageBase64' => true,
                ]);
                $qrcode = new QRCode($options);
                return $qrcode->render($conteudo);
            }
        } catch (\Throwable $e) {
            Log::warning("[QRCODE] Falha ao gerar via chillerlan: " . $e->getMessage());
        }

        // Gerador nativo de matriz QR Code em SVG vetorial (zero dependências externas)
        return QrCodeGenerator::renderDataUri($conteudo, 160);
    }

    /**
     * Obtém a lista de indexadores disponíveis na API
     */
    public function listarIndices(): array
    {
        try {
            $response = Http::timeout(5)
                ->withHeaders(['X-API-Key' => $this->apiKey])
                ->get("{$this->baseUrl}/v1/calculos/indices");

            if ($response->successful()) {
                return $response->json('indices') ?? [];
            }
        } catch (\Throwable $e) {
            Log::warning("[CALCULO_JUDICIAL] Falha ao listar índices: " . $e->getMessage());
        }

        return [
            ['codigo' => 'TJSP', 'nome' => 'Tabela Prática TJSP (Geral / Lei 14.905)'],
            ['codigo' => 'IPCA', 'nome' => 'IPCA (IBGE - Padrão Código Civil)'],
            ['codigo' => 'INPC', 'nome' => 'INPC (IBGE - Previdenciário / Trabalhista)'],
            ['codigo' => 'SELIC', 'nome' => 'Taxa SELIC Acumulada Mensal (% a.m.)'],
            ['codigo' => 'TAXA_LEGAL', 'nome' => 'Taxa Legal de Juros (Lei 14.905/2024 - Bacen)'],
        ];
    }

    /**
     * Decodifica o número CNJ sem custo e identifica o Tribunal
     */
    public function decodificarCNJ(string $numero): ?array
    {
        try {
            $response = Http::timeout(5)
                ->withHeaders(['X-API-Key' => $this->apiKey])
                ->get("{$this->baseUrl}/v1/cnj/" . urlencode($numero));

            if ($response->successful()) {
                return $response->json();
            }
        } catch (\Throwable $e) {
            Log::warning("[CALCULO_JUDICIAL] Falha ao decodificar CNJ: " . $e->getMessage());
        }

        return null;
    }

    /**
     * Liquidação Automática em 1 Clique:
     * Decodifica o CNJ, mapeia o indexador oficial do Tribunal e busca dados de autuação no DataJud (se disponível)
     */
    public function autocompletarPorProcesso(string $numero): array
    {
        $cnjInfo = $this->decodificarCNJ($numero);
        if (!$cnjInfo || empty($cnjInfo['valido'])) {
            return [
                'ok' => false,
                'mensagem' => 'Número de processo CNJ inválido ou dígito verificador incorreto.',
            ];
        }

        $tribunal = strtoupper($cnjInfo['tribunal'] ?? '');
        $indexadorSugerido = $this->mapearTribunalParaIndexador($tribunal);
        $anoProcesso = (int) ($cnjInfo['ano'] ?? date('Y'));
        $dataSugerida = sprintf('%04d-01-15', max(2023, min((int)date('Y'), $anoProcesso)));
        $valorSugerido = null;
        $descricaoSugerida = "Valor da Causa / Condenação ({$tribunal})";
        $classeProcessual = null;

        // Tenta consultar metadados rápidos do processo no DataJud (com timeout curto de 4s para não travar a UI)
        try {
            $respProc = Http::timeout(4)
                ->withHeaders(['X-API-Key' => $this->apiKey])
                ->get("{$this->baseUrl}/v1/processos/" . urlencode($cnjInfo['numero']) . "?modo=rapido&publicacoes=false");

            if ($respProc->successful()) {
                $proc = $respProc->json('processo') ?? [];
                if (!empty($proc['data_ajuizamento'])) {
                    $dtAjuiz = substr($proc['data_ajuizamento'], 0, 10);
                    if (preg_match('/^\d{4}-\d{2}-\d{2}$/', $dtAjuiz)) {
                        $dataSugerida = $dtAjuiz;
                    }
                }
                if (!empty($proc['valor_causa']) && is_numeric($proc['valor_causa'])) {
                    $valorSugerido = (float) $proc['valor_causa'];
                }
                if (!empty($proc['classe'])) {
                    $classeProcessual = $proc['classe'];
                    $descricaoSugerida = "{$proc['classe']} ({$tribunal})";
                }
            }
        } catch (\Throwable $e) {
            // Silencioso: mantém os dados decodificados offline do CNJ
        }

        return [
            'ok' => true,
            'numero_formatado' => $cnjInfo['numero'],
            'tribunal' => $tribunal,
            'segmento' => $cnjInfo['segmento'] ?? '',
            'ano' => $anoProcesso,
            'indexador_sugerido' => $indexadorSugerido,
            'data_sugerida' => $dataSugerida,
            'valor_sugerido' => $valorSugerido,
            'descricao_sugerida' => $descricaoSugerida,
            'classe' => $classeProcessual,
        ];
    }

    /**
     * Mapeia a sigla do Tribunal CNJ para o código de indexador correspondente na Calculadora
     */
    public function mapearTribunalParaIndexador(string $tribunal): string
    {
        $mapaDireto = [
            'TJSP' => 'TJSP',
            'TJRJ' => 'TJRJ',
            'TJMG' => 'TJMG',
            'TJRS' => 'TJRS',
            'TJPR' => 'TJPR',
            'TJSC' => 'TJSC',
            'TJES' => 'TJES',
            'TJCE' => 'TJCE',
            'TJDFT' => 'TJDF',
            'TJDF' => 'TJDF',
        ];

        if (isset($mapaDireto[$tribunal])) {
            return $mapaDireto[$tribunal];
        }

        if (str_starts_with($tribunal, 'TRF') || $tribunal === 'CJF' || $tribunal === 'STJ') {
            return 'JF_GERAL';
        }

        if (str_starts_with($tribunal, 'TRT') || $tribunal === 'TST') {
            return 'TST_TRABALHISTA';
        }

        if (str_starts_with($tribunal, 'TJ')) {
            return 'INPC';
        }

        return 'IPCA';
    }
}
