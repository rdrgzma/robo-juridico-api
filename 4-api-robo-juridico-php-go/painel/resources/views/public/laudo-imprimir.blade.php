<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Laudo Pericial de Liquidação Judicial #{{ substr($calculo->uuid, 0, 8) }}</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Cinzel:wght@600;700&family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        body { font-family: 'Inter', sans-serif; }
        .font-serif-legal { font-family: 'Cinzel', serif; }
        .font-mono-code { font-family: 'JetBrains Mono', monospace; }

        @media print {
            .no-print { display: none !important; }
            body { background: white !important; padding: 0 !important; color: black !important; }
            .page-break { page-break-after: always; }
            .laudo-container { box-shadow: none !important; border: none !important; max-width: 100% !important; margin: 0 !important; padding: 0 !important; }
            @page {
                size: A4 portrait;
                margin: 12mm 15mm 15mm 15mm;
            }
        }
    </style>
</head>
<body class="bg-slate-100 min-h-screen py-8 text-slate-900 antialiased">

    <!-- Barra de Controle Superior (Não aparece na impressão) -->
    <div class="no-print max-w-4xl mx-auto mb-6 px-4 flex flex-col sm:flex-row items-center justify-between gap-4 bg-slate-900 text-white p-4 rounded-2xl shadow-lg">
        <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-indigo-500/20 text-indigo-400 flex items-center justify-center font-bold text-xl border border-indigo-400/30">
                §
            </div>
            <div>
                <h1 class="text-sm font-bold text-white leading-tight">Robô Jurídico — Certificado de Liquidação</h1>
                <p class="text-xs text-slate-400">UUID: {{ $calculo->uuid }}</p>
            </div>
        </div>
        <div class="flex items-center gap-3">
            <a href="{{ route('laudo.exportar', ['uuid' => $calculo->uuid, 'formato' => 'xlsx']) }}" class="px-4 py-2 bg-emerald-700 hover:bg-emerald-600 text-xs font-semibold rounded-xl text-white transition-all flex items-center gap-1.5">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                </svg>
                Baixar Excel (.xlsx)
            </a>
            <a href="{{ route('validar.calculo', $calculo->uuid) }}" target="_blank" class="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-xs font-semibold rounded-xl text-slate-200 transition-all border border-slate-700 flex items-center gap-1.5">
                <svg class="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
                </svg>
                Verificar Página Pública
            </a>
            <button onclick="window.print()" class="px-5 py-2 bg-indigo-600 hover:bg-indigo-500 text-xs font-bold rounded-xl text-white shadow-md transition-all flex items-center gap-2">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z"/>
                </svg>
                Imprimir / Salvar em PDF
            </button>
        </div>
    </div>

    <!-- Container da Folha A4 do Laudo -->
    <div class="laudo-container max-w-4xl mx-auto bg-white border border-slate-200 shadow-xl rounded-2xl p-10 sm:p-12 relative overflow-hidden">
        
        <!-- Tarja de Segurança Superior -->
        <div class="absolute top-0 left-0 right-0 h-2 bg-gradient-to-r from-slate-800 via-indigo-900 to-slate-800"></div>

        <!-- Cabeçalho Oficial -->
        <div class="border-b-2 border-slate-900 pb-6 mb-6">
            <div class="flex items-start justify-between gap-6">
                <div class="space-y-1">
                    <div class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded text-[10px] font-bold tracking-widest bg-slate-900 text-white uppercase">
                        DOCUMENTO PERICIAL JUDICIAL
                    </div>
                    <h2 class="font-serif-legal text-2xl font-bold tracking-tight text-slate-900">
                        MEMÓRIA DISCRIMINADA DE CÁLCULO
                    </h2>
                    <p class="text-xs text-slate-600">
                        Liquidação de Débito e Atualização Monetária em Conformidade com o CPC/2015 e a Lei Federal nº 14.905/2024
                    </p>
                </div>

                <!-- QR Code de Autenticidade -->
                <div class="text-center shrink-0 flex flex-col items-center">
                    <div class="w-24 h-24 p-1 bg-white border-2 border-slate-900 rounded-lg shadow-sm">
                        <img src="{{ $qrCodeDataUri }}" alt="QR Code de Validação" class="w-full h-full object-contain">
                    </div>
                    <span class="text-[9px] font-bold text-slate-700 uppercase tracking-tighter mt-1">Autenticidade Digital</span>
                    <span class="text-[8px] text-slate-500 font-mono">Aponte a câmera</span>
                </div>
            </div>
        </div>

        <!-- Metadados do Processo e Solicitante -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 bg-slate-50 p-4 rounded-xl border border-slate-200 mb-6 text-xs">
            <div>
                <span class="text-slate-400 font-semibold uppercase text-[10px] block">Processo Judicial</span>
                <span class="font-bold text-slate-900 font-mono-code text-sm">
                    {{ $calculo->numero_processo ?: 'EXTRAJUDICIAL' }}
                </span>
            </div>
            <div>
                <span class="text-slate-400 font-semibold uppercase text-[10px] block">Data-Base de Atualização</span>
                <span class="font-bold text-slate-900">
                    {{ \Carbon\Carbon::parse($calculo->data_atualizacao)->format('d/m/Y') }}
                </span>
            </div>
            <div>
                <span class="text-slate-400 font-semibold uppercase text-[10px] block">Indexador de Correção</span>
                <span class="font-bold text-slate-900">
                    {{ $calculo->indexador_codigo }}
                </span>
            </div>
            <div>
                <span class="text-slate-400 font-semibold uppercase text-[10px] block">Emissão / Certificação</span>
                <span class="font-bold text-slate-900">
                    {{ $calculo->created_at->format('d/m/Y H:i') }}
                </span>
            </div>

            @if($calculo->lead_nome || $calculo->lead_oab)
                <div class="col-span-2 sm:col-span-4 pt-2 border-t border-slate-200 flex flex-wrap items-center justify-between gap-2 text-slate-600">
                    <div>
                        <span class="font-semibold text-slate-800">Profissional Solicitante:</span> 
                        {{ $calculo->lead_nome }} 
                        @if($calculo->lead_oab) &bull; <span class="font-mono">{{ $calculo->lead_oab }}</span> @endif
                    </div>
                    @if($calculo->lead_email)
                        <div class="text-slate-500 text-[11px]">{{ $calculo->lead_email }}</div>
                    @endif
                </div>
            @endif
        </div>

        <!-- Critérios Legais Aplicados -->
        <div class="mb-6">
            <h3 class="text-xs font-bold text-slate-900 uppercase tracking-wider mb-2 flex items-center gap-2">
                <span class="w-1.5 h-3 bg-indigo-600 rounded-sm"></span>
                1. Critérios Jurídicos e Contábeis Adotados
            </h3>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs bg-white border border-slate-200 rounded-xl p-3">
                <div class="space-y-0.5">
                    <span class="text-slate-500 font-medium block">Correção Monetária:</span>
                    <p class="font-semibold text-slate-800">
                        {{ $calculo->indexador_codigo }}
                    </p>
                    <span class="text-[10px] text-slate-400">
                        Deflação negativa: {{ !empty($requisicao['aplicar_deflacao']) ? 'Aplicada' : 'Não aplicada (preserva principal)' }}
                    </span>
                </div>

                <div class="space-y-0.5">
                    <span class="text-slate-500 font-medium block">Juros Moratórios:</span>
                    <p class="font-semibold text-slate-800">
                        @php
                            $jt = $requisicao['juros_mora']['tipo'] ?? ($requisicao['juros_tipo'] ?? 'CONVENCIONAL_MENSAL');
                            $jm = $requisicao['juros_mora']['modo'] ?? ($requisicao['juros_modo'] ?? 'SIMPLES');
                            $jp = $requisicao['juros_mora']['percentual_mensal'] ?? ($requisicao['juros_percentual'] ?? 1.0);
                        @endphp
                        @if($jt === 'LEI_14905' || $jt === 'TAXA_LEGAL')
                            Taxa Legal (Lei 14.905/2024 - SELIC menos IPCA)
                        @else
                            {{ number_format((float)$jp, 2, ',', '.') }}% a.m. (Juros {{ ucfirst(strtolower($jm)) }})
                        @endif
                    </p>
                    <span class="text-[10px] text-slate-400">Pro rata die diário ativo</span>
                </div>

                <div class="space-y-0.5">
                    <span class="text-slate-500 font-medium block">Penalidades Processuais:</span>
                    <p class="font-semibold text-slate-800">
                        @php
                            $fase = $requisicao['fase_execucao'] ?? $requisicao;
                            $penalidades = [];
                            if (!empty($fase['aplicar_art_523_multa'])) $penalidades[] = 'Multa 523 (10%)';
                            if (!empty($fase['aplicar_art_523_honorarios'])) $penalidades[] = 'Hon. 523 (10%)';
                            if (!empty($fase['multa_moratoria_percentual'])) $penalidades[] = 'Multa Moratória ' . $fase['multa_moratoria_percentual'] . '%';
                            if (!empty($fase['honorarios_sucumbenciais_percentual'])) $penalidades[] = 'Hon. Sucumb. ' . $fase['honorarios_sucumbenciais_percentual'] . '%';
                        @endphp
                        {{ count($penalidades) > 0 ? implode(' + ', $penalidades) : 'Sem encargos adicionais de execução' }}
                    </p>
                    <span class="text-[10px] text-slate-400">Fase de Cumprimento de Sentença</span>
                </div>
            </div>
        </div>

        <!-- Tabela Analítica de Parcelas -->
        <div class="mb-6">
            <h3 class="text-xs font-bold text-slate-900 uppercase tracking-wider mb-2 flex items-center gap-2">
                <span class="w-1.5 h-3 bg-indigo-600 rounded-sm"></span>
                2. Planilha Analítica das Parcelas Liquidadas
            </h3>
            
            <div class="overflow-x-auto border border-slate-200 rounded-xl">
                <table class="w-full text-left text-xs border-collapse">
                    <thead>
                        <tr class="bg-slate-100 text-slate-700 font-bold border-b border-slate-200">
                            <th class="py-2.5 px-3 w-10 text-center">Item</th>
                            <th class="py-2.5 px-3">Descrição / Natureza</th>
                            <th class="py-2.5 px-3">Vencimento</th>
                            <th class="py-2.5 px-3 text-right">Valor Nominal</th>
                            <th class="py-2.5 px-3 text-right">Fator Corr.</th>
                            <th class="py-2.5 px-3 text-right">Principal Corr.</th>
                            <th class="py-2.5 px-3 text-right">Juros</th>
                            <th class="py-2.5 px-3 text-right">Multa</th>
                            <th class="py-2.5 px-3 text-right bg-slate-200/60 font-black">Subtotal (R$)</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100 font-mono-code text-[11px]">
                        @php
                            $itensLaudo = $resultado['itens'] ?? $resultado['itens_liquidados'] ?? [];
                        @endphp
                        @foreach($itensLaudo as $it)
                            @php
                                $dtVencStr = $it['data_vencimento'] ?? '';
                                if (preg_match('/^\d{4}-\d{2}-\d{2}/', $dtVencStr)) {
                                    $dtVencStr = \Carbon\Carbon::parse($dtVencStr)->format('d/m/Y');
                                }
                                $valJurosItem = (float)($it['valor_juros'] ?? $it['valor_juros_mora'] ?? 0);
                                $valSubtotalItem = (float)($it['subtotal'] ?? $it['total_item'] ?? 0);
                            @endphp
                            <tr class="hover:bg-slate-50">
                                <td class="py-2 px-3 text-center text-slate-400 font-normal">{{ $it['numero_item'] }}</td>
                                <td class="py-2 px-3 font-medium font-sans text-slate-800">{{ $it['descricao'] }}</td>
                                <td class="py-2 px-3 text-slate-600">{{ $dtVencStr }}</td>
                                <td class="py-2 px-3 text-right text-slate-700">{{ number_format((float)$it['valor_original'], 2, ',', '.') }}</td>
                                <td class="py-2 px-3 text-right text-slate-500">{{ number_format((float)$it['fator_correcao'], 6, ',', '.') }}</td>
                                <td class="py-2 px-3 text-right font-semibold text-slate-900">{{ number_format((float)$it['valor_corrigido'], 2, ',', '.') }}</td>
                                <td class="py-2 px-3 text-right text-indigo-700 font-semibold">{{ number_format($valJurosItem, 2, ',', '.') }}</td>
                                <td class="py-2 px-3 text-right text-amber-700">{{ number_format((float)($it['valor_multa_moratoria'] ?? 0), 2, ',', '.') }}</td>
                                <td class="py-2 px-3 text-right font-black text-slate-950 bg-slate-50">
                                    {{ number_format($valSubtotalItem, 2, ',', '.') }}
                                </td>
                            </tr>
                        @endforeach
                    </tbody>
                </table>
            </div>
        </div>

        <!-- Quadro Resumo de Fechamento -->
        @php
            $resumoLaudo = $resultado['resumo'] ?? [];
            $notasLaudo = $resultado['notas_legais'] ?? $resultado['notas_metodologicas'] ?? [];
            $totalJurosLaudo = (float)($resumoLaudo['total_juros'] ?? $resumoLaudo['total_juros_mora'] ?? 0);
            $totalMultasLaudo = (float)($resumoLaudo['total_multas'] ?? (($resumoLaudo['total_multa_moratoria'] ?? 0) + ($resumoLaudo['multa_art_523_cpc'] ?? 0)));
            $totalHonLaudo = (float)($resumoLaudo['total_honorarios'] ?? (($resumoLaudo['honorarios_art_523_cpc'] ?? 0) + ($resumoLaudo['honorarios_sucumbenciais'] ?? 0)));
            $totalGeralLaudo = (float)($resumoLaudo['total_geral'] ?? $resumoLaudo['total_geral_devido'] ?? 0);
        @endphp
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-6 mb-8">
            <!-- Notas Técnicas / Legais -->
            <div class="bg-slate-50 border border-slate-200 rounded-xl p-4 text-xs space-y-2">
                <span class="font-bold text-slate-900 uppercase tracking-wider text-[10px] block">Fundamentação e Notas Técnicas</span>
                <ul class="list-disc list-inside text-slate-600 space-y-1 text-[11px] leading-relaxed">
                    @forelse($notasLaudo as $nota)
                        <li>{{ $nota }}</li>
                    @empty
                        <li>Liquidação executada de acordo com as normas consolidadas da Corregedoria Geral da Justiça e Código de Processo Civil.</li>
                        <li>Os fatores de correção contemplam a variação integral acumulada do indexador até a data-base indicada.</li>
                    @endforelse
                </ul>
            </div>

            <!-- Totais Finais -->
            <div class="bg-slate-900 text-white rounded-xl p-5 shadow-inner">
                <span class="text-slate-400 font-bold uppercase tracking-wider text-[10px] block mb-3">Demonstrativo Consolidado</span>
                <div class="space-y-2 text-xs border-b border-slate-800 pb-3 font-mono-code">
                    <div class="flex justify-between">
                        <span class="text-slate-400 font-sans">Principal Corrigido:</span>
                        <span class="font-semibold">R$ {{ number_format((float)($resumoLaudo['total_principal_corrigido'] ?? 0), 2, ',', '.') }}</span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-slate-400 font-sans">Juros Moratórios Acumulados:</span>
                        <span class="font-semibold text-indigo-300">R$ {{ number_format($totalJurosLaudo, 2, ',', '.') }}</span>
                    </div>
                    @if($totalMultasLaudo > 0)
                        <div class="flex justify-between">
                            <span class="text-slate-400 font-sans">Multas Processuais (Mora / Art. 523):</span>
                            <span class="font-semibold text-amber-300">R$ {{ number_format($totalMultasLaudo, 2, ',', '.') }}</span>
                        </div>
                    @endif
                    @if($totalHonLaudo > 0)
                        <div class="flex justify-between">
                            <span class="text-slate-400 font-sans">Honorários Advocatícios:</span>
                            <span class="font-semibold text-emerald-300">R$ {{ number_format($totalHonLaudo, 2, ',', '.') }}</span>
                        </div>
                    @endif
                </div>

                <div class="pt-3 flex justify-between items-baseline">
                    <span class="font-bold text-sm tracking-tight font-sans">VALOR TOTAL DA EXECUÇÃO:</span>
                    <span class="text-2xl font-black text-emerald-400 font-mono-code">
                        R$ {{ number_format($totalGeralLaudo, 2, ',', '.') }}
                    </span>
                </div>
            </div>
        </div>

        <!-- Rodapé de Autenticidade e Assinatura Digital -->
        <div class="border-t-2 border-slate-900 pt-4 text-[10px] text-slate-600 space-y-2">
            <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2">
                <div>
                    <span class="font-bold text-slate-900 uppercase">Validação Pública de Não-Adulteração:</span>
                    <span class="font-mono text-slate-800">{{ $urlValidacao }}</span>
                </div>
                <div class="text-right font-mono text-[9px] text-slate-500">
                    UUID: {{ $calculo->uuid }}
                </div>
            </div>

            <div class="bg-slate-100 p-2.5 rounded border border-slate-200 font-mono text-[9px] text-slate-700 break-all select-all">
                <span class="font-bold text-slate-900 font-sans">HASH CRIPTOGRÁFICO SHA-256 DO DOCUMENTO:</span> 
                {{ $calculo->hash_validacao }}
            </div>

            <div class="flex items-center justify-between text-slate-400 text-[9px] pt-1">
                <span>Robô Jurídico &copy; {{ date('Y') }} — Motor de Cálculo Judicial de Alta Precisão</span>
                <span>Página 1 de 1</span>
            </div>
        </div>

    </div>

</body>
</html>
