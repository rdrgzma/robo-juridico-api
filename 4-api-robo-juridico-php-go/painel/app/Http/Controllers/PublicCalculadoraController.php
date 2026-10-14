<?php

namespace App\Http\Controllers;

use App\Services\CalculoJudicialService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\View\View;

class PublicCalculadoraController extends Controller
{
    protected CalculoJudicialService $service;

    public function __construct(CalculoJudicialService $service)
    {
        $this->service = $service;
    }

    /**
     * Exibe a página pública de cálculo judicial (Lead Magnet)
     */
    public function index(): View
    {
        $indices = $this->service->listarIndices();
        return view('public.calculadora', compact('indices'));
    }

    /**
     * Executa o cálculo em tempo real via AJAX
     */
    public function calcular(Request $request): JsonResponse
    {
        $dados = $request->validate([
            'numero_processo' => 'nullable|string',
            'data_atualizacao' => 'required|date',
            'indexador_codigo' => 'required|string',
            'aplicar_deflacao' => 'nullable|boolean',
            'juros_mora' => 'nullable|array',
            'fase_execucao' => 'nullable|array',
            'parcelas' => 'required|array|min:1',
            'parcelas.*.data_vencimento' => 'required|date',
            'parcelas.*.valor_original' => 'required|numeric',
            'parcelas.*.descricao' => 'nullable|string',
            'parcelas.*.tipo' => 'nullable|string',
        ]);

        $resultado = $this->service->calcular($dados);

        return response()->json($resultado);
    }

    /**
     * Importa arquivo Excel (.xlsx, .csv) ou texto colado via Go Engine
     */
    public function importarPlanilha(Request $request): JsonResponse
    {
        if ($request->hasFile('planilha')) {
            $resultado = $this->service->importarPlanilha($request->file('planilha'));
        } elseif ($request->filled('texto')) {
            $resultado = $this->service->importarPlanilha($request->input('texto'), 'dados_colados.csv');
        } else {
            return response()->json([
                'ok' => false,
                'mensagem' => 'Nenhum arquivo ou texto colado foi fornecido.',
            ], 422);
        }

        return response()->json($resultado);
    }

    /**
     * Captura os dados do Lead e salva a liquidação gerando Certificado com Hash
     */
    public function emitirLaudo(Request $request): JsonResponse
    {
        $request->validate([
            'nome' => 'required|string|max:150',
            'email' => 'required|email|max:150',
            'whatsapp' => 'required|string|max:30',
            'oab' => 'nullable|string|max:30',
            'dados_calculo' => 'required|array',
            'resultado' => 'required|array',
        ]);

        $leadData = [
            'nome' => $request->input('nome'),
            'email' => $request->input('email'),
            'whatsapp' => $request->input('whatsapp'),
            'oab' => $request->input('oab'),
        ];

        $calculo = $this->service->salvarCalculoComLead(
            $request->input('dados_calculo'),
            $request->input('resultado'),
            $leadData,
            $request->ip()
        );

        return response()->json([
            'ok' => true,
            'uuid' => $calculo->uuid,
            'hash' => $calculo->hash_validacao,
            'url_laudo' => route('laudo.imprimir', $calculo->uuid),
            'url_validacao' => route('validar.calculo', $calculo->uuid),
        ]);
    }

    /**
     * Exibe o Laudo Pericial Certificado para impressão direta / geração de PDF Nativa
     */
    public function imprimirLaudo(string $uuid): View
    {
        $calculo = $this->service->buscarPorUuid($uuid);
        if (!$calculo) {
            abort(404, 'Laudo Pericial não encontrado.');
        }

        $urlValidacao = route('validar.calculo', $calculo->uuid);
        $qrCodeDataUri = $this->service->gerarQrCodeDataUri($urlValidacao);

        return view('public.laudo-imprimir', [
            'calculo' => $calculo,
            'urlValidacao' => $urlValidacao,
            'qrCodeDataUri' => $qrCodeDataUri,
            'resultado' => $calculo->resultado_json,
            'requisicao' => $calculo->payload_requisicao,
        ]);
    }

    /**
     * Página Pública de Validação de Autenticidade (Selo de Integridade)
     */
    public function validar(string $uuid): View
    {
        $autenticidade = $this->service->verificarAutenticidade($uuid);

        return view('public.validar-laudo', [
            'valido' => $autenticidade['valido'],
            'motivo' => $autenticidade['motivo'] ?? null,
            'calculo' => $autenticidade['calculo'] ?? null,
            'hashValidacao' => $autenticidade['hash_validacao'] ?? '',
            'dataEmissao' => $autenticidade['data_emissao'] ?? '',
        ]);
    }
}
