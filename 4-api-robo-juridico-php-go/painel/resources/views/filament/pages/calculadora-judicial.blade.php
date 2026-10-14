<x-filament-panels::page>
    <div class="space-y-6">
        {{-- Formulário de Parâmetros e Lançamentos --}}
        <div class="p-6 bg-white rounded-xl shadow-sm border border-gray-200 dark:bg-gray-900 dark:border-gray-800 print:hidden">
            <form wire:submit.prevent="calcular">
                {{ $this->form }}

                <div class="flex items-center gap-4 mt-6">
                    <x-filament::button type="submit" size="lg" icon="heroicon-m-calculator" color="primary">
                        Executar Liquidação Judicial
                    </x-filament::button>

                    <x-filament::button wire:click="limpar" color="gray" type="button" icon="heroicon-m-arrow-path">
                        Limpar Formulário
                    </x-filament::button>

                    @if($resultado)
                        <x-filament::button wire:click="emitirLaudoPdf" color="success" type="button" icon="heroicon-m-shield-check">
                            Emitir Laudo Certificado com QR Code (PDF)
                        </x-filament::button>

                        <x-filament::button onclick="window.print()" color="gray" type="button" icon="heroicon-m-printer">
                            Imprimir Tela
                        </x-filament::button>
                    @endif
                </div>
            </form>
        </div>

        {{-- Exibição do Resultado da Liquidação --}}
        @if($resultado && !empty($resultado['sucesso']))
            @php
                $resumo = $resultado['resumo'];
                $itens = $resultado['itens_liquidados'];
                $notas = $resultado['notas_metodologicas'] ?? [];
            @endphp

            <div id="laudo-pericial" class="space-y-6">
                {{-- Cabeçalho Formal para Impressão --}}
                <div class="p-6 bg-white rounded-xl shadow-sm border border-gray-200 dark:bg-gray-900 dark:border-gray-800">
                    <div class="flex flex-col md:flex-row justify-between items-start md:items-center border-b pb-4 mb-4 border-gray-200 dark:border-gray-700">
                        <div>
                            <h2 class="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
                                Demonstrativo de Débito e Liquidação Judicial
                            </h2>
                            <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                                Memória de cálculo discriminada e atualizada com base nas normas processuais e legais vigentes.
                            </p>
                        </div>
                        <div class="mt-4 md:mt-0 text-right">
                            <span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
                                Cálculo Válido
                            </span>
                            <p class="text-xs text-gray-400 mt-1">Data-base: {{ $resumo['data_atualizacao'] }}</p>
                        </div>
                    </div>

                    {{-- Cards de Resumo Sintético (KPIs) --}}
                    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
                        <div class="p-4 bg-gray-50 rounded-lg dark:bg-gray-800">
                            <span class="text-xs text-gray-500 uppercase tracking-wider font-semibold">Principal Original</span>
                            <div class="text-lg font-bold text-gray-900 dark:text-white mt-1">
                                R$ {{ number_format((float)$resumo['total_principal_original'], 2, ',', '.') }}
                            </div>
                        </div>

                        <div class="p-4 bg-gray-50 rounded-lg dark:bg-gray-800">
                            <span class="text-xs text-gray-500 uppercase tracking-wider font-semibold">Correção Monetária</span>
                            <div class="text-lg font-bold text-blue-600 dark:text-blue-400 mt-1">
                                + R$ {{ number_format((float)$resumo['total_correcao_monetaria'], 2, ',', '.') }}
                            </div>
                        </div>

                        <div class="p-4 bg-gray-50 rounded-lg dark:bg-gray-800">
                            <span class="text-xs text-gray-500 uppercase tracking-wider font-semibold">Juros de Mora</span>
                            <div class="text-lg font-bold text-indigo-600 dark:text-indigo-400 mt-1">
                                + R$ {{ number_format((float)$resumo['total_juros_mora'], 2, ',', '.') }}
                            </div>
                        </div>

                        <div class="p-4 bg-emerald-50 rounded-lg dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800">
                            <span class="text-xs text-emerald-800 dark:text-emerald-300 uppercase tracking-wider font-semibold">Total Geral Devido</span>
                            <div class="text-xl font-black text-emerald-600 dark:text-emerald-400 mt-1">
                                R$ {{ number_format((float)$resumo['total_geral_devido'], 2, ',', '.') }}
                            </div>
                        </div>
                    </div>
                </div>

                {{-- Tabela Analítica de Liquidação Item a Item --}}
                <div class="p-6 bg-white rounded-xl shadow-sm border border-gray-200 dark:bg-gray-900 dark:border-gray-800 overflow-x-auto">
                    <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-4">
                        Detalhamento Analítico das Parcelas
                    </h3>

                    <table class="w-full text-left text-sm text-gray-600 dark:text-gray-300 border-collapse">
                        <thead>
                            <tr class="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/60 text-xs uppercase font-bold text-gray-500 dark:text-gray-400">
                                <th class="p-3">#</th>
                                <th class="p-3">Descrição</th>
                                <th class="p-3">Natureza</th>
                                <th class="p-3">Vencimento</th>
                                <th class="p-3 text-right">Valor Original</th>
                                <th class="p-3 text-right">Fator</th>
                                <th class="p-3 text-right">Valor Corrigido</th>
                                <th class="p-3 text-right">Taxa Juros</th>
                                <th class="p-3 text-right">Juros (R$)</th>
                                <th class="p-3 text-right">Multa (R$)</th>
                                <th class="p-3 text-right">Total Item (R$)</th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                            @foreach($itens as $item)
                                <tr class="hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
                                    <td class="p-3 font-medium">{{ $item['numero_item'] }}</td>
                                    <td class="p-3 font-semibold text-gray-900 dark:text-white">{{ $item['descricao'] }}</td>
                                    <td class="p-3">
                                        <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium 
                                            {{ $item['tipo'] === 'DEBITO' ? 'bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-300' : '' }}
                                            {{ $item['tipo'] === 'CUSTA_PROCESSUAL' ? 'bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300' : '' }}
                                            {{ $item['tipo'] === 'PAGAMENTO_PARCIAL' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-300' : '' }}">
                                            {{ $item['tipo'] }}
                                        </span>
                                    </td>
                                    <td class="p-3">{{ $item['data_vencimento'] }}</td>
                                    <td class="p-3 text-right">R$ {{ number_format((float)$item['valor_original'], 2, ',', '.') }}</td>
                                    <td class="p-3 text-right font-mono text-xs">{{ number_format((float)$item['fator_correcao'], 6, ',', '.') }}</td>
                                    <td class="p-3 text-right font-semibold">R$ {{ number_format((float)$item['valor_corrigido'], 2, ',', '.') }}</td>
                                    <td class="p-3 text-right">{{ number_format((float)$item['taxa_juros_acumulada_percentual'], 2, ',', '.') }}%</td>
                                    <td class="p-3 text-right">R$ {{ number_format((float)$item['valor_juros_mora'], 2, ',', '.') }}</td>
                                    <td class="p-3 text-right">R$ {{ number_format((float)$item['valor_multa_moratoria'], 2, ',', '.') }}</td>
                                    <td class="p-3 text-right font-bold text-gray-900 dark:text-white">R$ {{ number_format((float)$item['total_item'], 2, ',', '.') }}</td>
                                </tr>
                            @endforeach
                        </tbody>
                        <tfoot class="border-t-2 border-gray-300 dark:border-gray-600 bg-gray-50/80 dark:bg-gray-800/80 font-bold">
                            <tr>
                                <td colspan="4" class="p-3 text-right">Subtotal das Parcelas:</td>
                                <td class="p-3 text-right">R$ {{ number_format((float)$resumo['total_principal_original'], 2, ',', '.') }}</td>
                                <td></td>
                                <td class="p-3 text-right">R$ {{ number_format((float)$resumo['total_principal_corrigido'], 2, ',', '.') }}</td>
                                <td></td>
                                <td class="p-3 text-right">R$ {{ number_format((float)$resumo['total_juros_mora'], 2, ',', '.') }}</td>
                                <td class="p-3 text-right">R$ {{ number_format((float)$resumo['total_multa_moratoria'], 2, ',', '.') }}</td>
                                <td class="p-3 text-right text-emerald-600 dark:text-emerald-400">R$ {{ number_format((float)$resumo['subtotal_condenacao'], 2, ',', '.') }}</td>
                            </tr>
                        </tfoot>
                    </table>

                    {{-- Fechamento Executivo (Art. 523 CPC e Honorários) --}}
                    <div class="mt-6 flex justify-end">
                        <div class="w-full md:w-96 space-y-2 border-t pt-4 border-gray-200 dark:border-gray-700">
                            <div class="flex justify-between text-sm">
                                <span class="text-gray-500">Subtotal da Condenação:</span>
                                <span class="font-semibold">R$ {{ number_format((float)$resumo['subtotal_condenacao'], 2, ',', '.') }}</span>
                            </div>

                            @if((float)$resumo['multa_art_523_cpc'] > 0)
                                <div class="flex justify-between text-sm text-red-600 dark:text-red-400">
                                    <span>Multa Art. 523, § 1º CPC (10%):</span>
                                    <span class="font-semibold">+ R$ {{ number_format((float)$resumo['multa_art_523_cpc'], 2, ',', '.') }}</span>
                                </div>
                            @endif

                            @if((float)$resumo['honorarios_art_523_cpc'] > 0)
                                <div class="flex justify-between text-sm text-indigo-600 dark:text-indigo-400">
                                    <span>Honorários Executivos Art. 523 CPC (10%):</span>
                                    <span class="font-semibold">+ R$ {{ number_format((float)$resumo['honorarios_art_523_cpc'], 2, ',', '.') }}</span>
                                </div>
                            @endif

                            @if((float)$resumo['honorarios_sucumbenciais'] > 0)
                                <div class="flex justify-between text-sm text-indigo-600 dark:text-indigo-400">
                                    <span>Honorários Sucumbenciais:</span>
                                    <span class="font-semibold">+ R$ {{ number_format((float)$resumo['honorarios_sucumbenciais'], 2, ',', '.') }}</span>
                                </div>
                            @endif

                            <div class="flex justify-between text-base font-black border-t-2 pt-2 border-gray-900 dark:border-white text-gray-900 dark:text-white">
                                <span>TOTAL GERAL DA LIQUIDAÇÃO:</span>
                                <span class="text-emerald-600 dark:text-emerald-400 text-lg">
                                    R$ {{ number_format((float)$resumo['total_geral_devido'], 2, ',', '.') }}
                                </span>
                            </div>
                        </div>
                    </div>
                </div>

                {{-- Notas Metodológicas e Fundamentação Legal --}}
                <div class="p-6 bg-white rounded-xl shadow-sm border border-gray-200 dark:bg-gray-900 dark:border-gray-800">
                    <h4 class="text-sm font-bold uppercase tracking-wider text-gray-500 mb-2">
                        Notas Metodológicas e Fundamentação Jurídica
                    </h4>
                    <ul class="list-disc list-inside text-xs text-gray-600 dark:text-gray-400 space-y-1">
                        @foreach($notas as $nota)
                            <li>{{ $nota }}</li>
                        @endforeach
                        <li>Calculado automaticamente pelo motor de alta precisão Robô Jurídico Calc Engine em conformidade com as Súmulas do STJ e CPC/2015.</li>
                    </ul>
                </div>
            </div>
        @endif
    </div>

    {{-- Estilos Específicos para Impressão de Laudo Pericial --}}
    <style>
        @media print {
            body { background: white !important; color: black !important; font-size: 11pt !important; }
            .fi-sidebar, .fi-topbar, .print\:hidden { display: none !important; }
            #laudo-pericial { display: block !important; width: 100% !important; margin: 0 !important; padding: 0 !important; }
            table { font-size: 9pt !important; }
        }
    </style>
</x-filament-panels::page>
