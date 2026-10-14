<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Certificação de Autenticidade | Robô Jurídico</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
        body { font-family: 'Inter', sans-serif; }
    </style>
</head>
<body class="bg-slate-50 min-h-screen text-slate-800 antialiased flex flex-col justify-between">
    <!-- Header -->
    <header class="bg-white border-b border-slate-200 shadow-sm py-4">
        <div class="max-w-4xl mx-auto px-4 flex items-center justify-between">
            <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-indigo-600 flex items-center justify-center text-white shadow-md font-bold text-xl">
                    §
                </div>
                <div>
                    <h1 class="text-lg font-bold text-slate-900 leading-tight">Robô Jurídico</h1>
                    <p class="text-xs text-slate-500 font-medium">Portal Público de Validação & Autenticidade Digital</p>
                </div>
            </div>
            <a href="{{ route('calculadora.publica') }}" class="text-xs font-semibold text-indigo-600 hover:text-indigo-800 flex items-center gap-1">
                &larr; Nova Liquidação
            </a>
        </div>
    </header>

    <!-- Main Container -->
    <main class="max-w-3xl mx-auto px-4 py-10 w-full">
        @if($valido)
            <!-- Card de Sucesso / Autenticidade Confirmada -->
            <div class="bg-white rounded-2xl shadow-xl border border-emerald-100 overflow-hidden">
                <div class="bg-gradient-to-r from-emerald-600 to-teal-700 px-8 py-6 text-white text-center sm:text-left sm:flex sm:items-center sm:justify-between">
                    <div class="flex items-center gap-4">
                        <div class="w-14 h-14 rounded-full bg-white/20 backdrop-blur flex items-center justify-center shrink-0 border border-white/30">
                            <svg class="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
                            </svg>
                        </div>
                        <div>
                            <span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/40 text-emerald-100 border border-emerald-400/30 uppercase tracking-wider mb-1">
                                Selo Digital Verificado
                            </span>
                            <h2 class="text-2xl font-bold tracking-tight">Cálculo Autêntico e Íntegro</h2>
                            <p class="text-emerald-100 text-xs mt-0.5">Os dados coincidem perfeitamente com a liquidação gerada e registrada no repositório.</p>
                        </div>
                    </div>
                </div>

                <div class="p-8 space-y-6">
                    <!-- Resumo Geral -->
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div class="bg-slate-50 p-4 rounded-xl border border-slate-100">
                            <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider block">Processo Judicial</span>
                            <span class="text-base font-bold text-slate-800">
                                {{ $calculo->numero_processo ?: 'Liquidação Extrajudicial / Pré-Processual' }}
                            </span>
                        </div>

                        <div class="bg-slate-50 p-4 rounded-xl border border-slate-100">
                            <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider block">Indexador Aplicado</span>
                            <span class="text-base font-bold text-slate-800">
                                {{ $calculo->indexador_codigo }} (Tabela Oficial)
                            </span>
                        </div>

                        <div class="bg-slate-50 p-4 rounded-xl border border-slate-100">
                            <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider block">Data-Base de Atualização</span>
                            <span class="text-base font-bold text-slate-800">
                                {{ \Carbon\Carbon::parse($calculo->data_atualizacao)->format('d/m/Y') }}
                            </span>
                        </div>

                        <div class="bg-slate-50 p-4 rounded-xl border border-slate-100">
                            <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider block">Data de Registro e Emissão</span>
                            <span class="text-base font-bold text-slate-800">
                                {{ $dataEmissao }}
                            </span>
                        </div>
                    </div>

                    <!-- Total do Cálculo -->
                    <div class="p-5 rounded-xl bg-emerald-50/60 border border-emerald-200 flex items-center justify-between">
                        <div>
                            <span class="text-xs font-semibold text-emerald-800 uppercase tracking-wider">Valor Total Atualizado e Certificado</span>
                            <p class="text-xs text-emerald-700">Inclui correção monetária, juros de mora e eventuais penalidades processuais.</p>
                        </div>
                        <div class="text-2xl font-black text-emerald-900 tracking-tight">
                            R$ {{ number_format($calculo->total_geral, 2, ',', '.') }}
                        </div>
                    </div>

                    <!-- Hash Criptográfico -->
                    <div class="bg-slate-900 text-slate-200 p-4 rounded-xl font-mono text-xs">
                        <div class="flex items-center justify-between mb-1.5 text-slate-400">
                            <span class="font-bold text-emerald-400 flex items-center gap-1.5">
                                <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                                HASH SHA-256 DE AUTENTICIDADE:
                            </span>
                            <span class="text-[10px] uppercase">À prova de adulteração</span>
                        </div>
                        <div class="break-all select-all text-slate-300 font-semibold bg-slate-800/80 p-2.5 rounded border border-slate-700">
                            {{ $hashValidacao }}
                        </div>
                        <p class="text-[11px] text-slate-400 mt-2">
                            Identificador Único (UUID): <span class="text-slate-300">{{ $calculo->uuid }}</span>
                        </p>
                    </div>

                    <!-- Ações -->
                    <div class="pt-4 border-t border-slate-200 flex flex-col sm:flex-row items-center justify-between gap-3">
                        <a href="{{ route('laudo.imprimir', $calculo->uuid) }}" target="_blank" class="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-6 py-3 rounded-xl font-bold text-white bg-indigo-600 hover:bg-indigo-700 shadow-md transition-all">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z"/>
                            </svg>
                            Visualizar / Imprimir Laudo Pericial (PDF)
                        </a>

                        <a href="{{ route('calculadora.publica') }}" class="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-5 py-3 rounded-xl font-medium text-slate-600 hover:text-slate-800 hover:bg-slate-100 transition-all text-sm">
                            Fazer Novo Cálculo
                        </a>
                    </div>
                </div>
            </div>
        @else
            <!-- Card de Erro / Não Encontrado ou Violado -->
            <div class="bg-white rounded-2xl shadow-xl border border-rose-100 overflow-hidden text-center p-10">
                <div class="w-16 h-16 rounded-full bg-rose-100 text-rose-600 mx-auto flex items-center justify-center mb-4">
                    <svg class="w-9 h-9" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
                    </svg>
                </div>
                <h2 class="text-2xl font-bold text-slate-900 mb-2">Autenticidade Não Confirmada</h2>
                <p class="text-slate-600 text-sm max-w-md mx-auto mb-6">
                    {{ $motivo ?? 'O identificador informado não corresponde a nenhum cálculo certificado em nosso repositório ou o laudo sofreu alterações não autorizadas.' }}
                </p>
                <a href="{{ route('calculadora.publica') }}" class="inline-flex items-center gap-2 px-6 py-2.5 rounded-xl font-semibold text-white bg-indigo-600 hover:bg-indigo-700 shadow transition-all">
                    Acessar Calculadora Judicial
                </a>
            </div>
        @endif
    </main>

    <!-- Footer -->
    <footer class="bg-white border-t border-slate-200 py-6 text-center text-xs text-slate-500">
        <p>&copy; {{ date('Y') }} Robô Jurídico Inteligente. Certificação Digital baseada em SHA-256 e Lei 14.905/2024.</p>
    </footer>
</body>
</html>
