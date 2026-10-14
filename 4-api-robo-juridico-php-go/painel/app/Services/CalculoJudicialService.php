<?php

namespace App\Services;

use App\Models\CalculoSalvo;
use chillerlan\QRCode\QRCode;
use chillerlan\QRCode\QROptions;
use Illuminate\Http\UploadedFile;
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

            $erro = $response->json('detalhe') ?? $response->json('mensagem') ?? $response->body();
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

            $erro = $response->json('detalhe') ?? $response->json('mensagem') ?? $response->body();
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
     * Salva o cálculo realizado vinculando ao Lead e gerando hash criptográfico de autenticidade
     */
    public function salvarCalculoComLead(
        array $dadosCalculo,
        array $resultado,
        array $leadData = [],
        ?string $ipOrigem = null
    ): CalculoSalvo {
        $uuid = Str::uuid()->toString();

        // Hash SHA-256 à prova de fraude do resultado e do processo
        $hashBase = json_encode([
            'uuid' => $uuid,
            'processo' => $dadosCalculo['numero_processo'] ?? '',
            'data_atualizacao' => $dadosCalculo['data_atualizacao'] ?? '',
            'total_geral' => $resultado['resumo']['total_geral'] ?? 0,
            'itens_count' => count($resultado['itens'] ?? []),
        ]);
        $hashValidacao = hash('sha256', $hashBase . 'ROBO_JURIDICO_SALT_2026');

        return CalculoSalvo::create([
            'uuid' => $uuid,
            'hash_validacao' => $hashValidacao,
            'numero_processo' => $dadosCalculo['numero_processo'] ?? null,
            'data_atualizacao' => $dadosCalculo['data_atualizacao'] ?? now()->format('Y-m-d'),
            'indexador_codigo' => $dadosCalculo['indexador_codigo'] ?? 'TJSP',
            'total_geral' => (float)($resultado['resumo']['total_geral'] ?? 0),
            'payload_requisicao' => $dadosCalculo,
            'resultado_json' => $resultado,
            'lead_nome' => $leadData['nome'] ?? null,
            'lead_email' => $leadData['email'] ?? null,
            'lead_whatsapp' => $leadData['whatsapp'] ?? null,
            'lead_oab' => $leadData['oab'] ?? null,
            'ip_origem' => $ipOrigem ?? request()->ip(),
        ]);
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

        $hashBase = json_encode([
            'uuid' => $calculo->uuid,
            'processo' => $dadosCalculo['numero_processo'] ?? '',
            'data_atualizacao' => $dadosCalculo['data_atualizacao'] ?? '',
            'total_geral' => $resultado['resumo']['total_geral'] ?? 0,
            'itens_count' => count($resultado['itens'] ?? []),
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
     * Gera QR Code em formato Data URI (SVG / PNG base64) para impressão no Laudo Pericial
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

        // Fallback dinâmico para renderização de SVG puro embutido
        return 'data:image/svg+xml;utf8,' . rawurlencode(
            '<svg xmlns="http://www.w3.org/2000/svg" width="120" height="120" viewBox="0 0 120 120">' .
            '<rect width="120" height="120" fill="#f8fafc" stroke="#cbd5e1" stroke-width="2"/>' .
            '<text x="60" y="55" font-size="11" font-family="sans-serif" text-anchor="middle" fill="#1e293b">VALIDAR LAUDO</text>' .
            '<text x="60" y="75" font-size="9" font-family="sans-serif" text-anchor="middle" fill="#0284c7">QR Code Seguro</text>' .
            '</svg>'
        );
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
}
