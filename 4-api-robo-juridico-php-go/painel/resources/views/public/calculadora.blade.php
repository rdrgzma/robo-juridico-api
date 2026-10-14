<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Calculadora e Liquidação Judicial Online | Robô Jurídico</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
    <style>
        body { font-family: 'Inter', sans-serif; }
        .font-mono-code { font-family: 'JetBrains Mono', monospace; }
        [x-cloak] { display: none !important; }
    </style>
</head>
<body class="bg-slate-50 min-h-screen text-slate-800 antialiased flex flex-col justify-between">

    <!-- Navbar Pública -->
    <header class="bg-white border-b border-slate-200 sticky top-0 z-30 shadow-xs">
        <div class="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
            <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-700 to-indigo-500 flex items-center justify-center text-white shadow-md font-bold text-xl">
                    §
                </div>
                <div>
                    <div class="flex items-center gap-2">
                        <span class="text-lg font-black text-slate-900 tracking-tight">Robô Jurídico</span>
                        <span class="px-2 py-0.5 rounded text-[10px] font-bold bg-indigo-50 text-indigo-700 border border-indigo-200">
                            LEI 14.905/2024
                        </span>
                    </div>
                    <p class="text-xs text-slate-500 font-medium hidden sm:block">Calculadora e Liquidação Judicial de Alta Precisão</p>
                </div>
            </div>

            <div class="flex items-center gap-3">
                <a href="/admin" class="text-xs font-semibold text-slate-600 hover:text-indigo-600 px-3 py-1.5 rounded-lg border border-slate-200 hover:border-indigo-300 transition-all">
                    Área do Assinante &rarr;
                </a>
            </div>
        </div>
    </header>

    <!-- Hero / Lead Banner -->
    <div class="bg-gradient-to-b from-indigo-950 via-slate-900 to-slate-900 text-white py-12 px-4 border-b border-slate-800">
        <div class="max-w-4xl mx-auto text-center space-y-4">
            <span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                Paridade 100% com DrCalc e Tabelas Práticas Oficiais
            </span>
            <h1 class="text-3xl sm:text-5xl font-extrabold tracking-tight">
                Calculadora e Liquidação Judicial
            </h1>
            <p class="text-slate-300 text-sm sm:text-base max-w-2xl mx-auto">
                Atualize débitos com a Tabela Prática de todos os Tribunais (TJSP, TJRJ, TJMG, TJRS, TJPR, TRF, TST), Taxa Legal da Lei 14.905/2024, juros com pro rata die e emissão de laudo certificado com QR Code.
            </p>
        </div>
    </div>

    <!-- Container Principal do Formulário e Cálculos -->
    <main class="max-w-6xl mx-auto px-4 -mt-6 mb-16 w-full space-y-8">
        
        <form id="formCalculo" onsubmit="event.preventDefault(); executarCalculo();" class="space-y-6">
            @csrf

            <!-- Card 1: Parâmetros Gerais e Tribunal -->
            <div class="bg-white rounded-2xl shadow-xl border border-slate-200 p-6 sm:p-8 space-y-6">
                <div class="flex items-center justify-between border-b border-slate-100 pb-4">
                    <h2 class="text-base font-bold text-slate-900 flex items-center gap-2">
                        <span class="w-6 h-6 rounded-lg bg-indigo-100 text-indigo-700 flex items-center justify-center text-xs font-black">1</span>
                        Identificação e Critérios da Liquidação
                    </h2>
                    <span class="text-xs text-slate-400">Preencha os parâmetros da ação</span>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-3 gap-5">
                    <!-- Número CNJ -->
                    <div>
                        <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                            Processo CNJ (Opcional)
                        </label>
                        <div class="relative">
                            <input type="text" id="numero_processo" name="numero_processo" placeholder="0000000-00.0000.0.00.0000" class="w-full text-xs font-mono rounded-xl border-slate-300 focus:border-indigo-500 focus:ring-indigo-500 p-2.5 border">
                        </div>
                        <span class="text-[10px] text-slate-400 mt-1 block">Insira para registrar no laudo pericial</span>
                    </div>

                    <!-- Indexador de Correção -->
                    <div>
                        <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                            Indexador de Correção Monetária *
                        </label>
                        <select id="indexador_codigo" name="indexador_codigo" required class="w-full text-xs rounded-xl border-slate-300 focus:border-indigo-500 focus:ring-indigo-500 p-2.5 border font-medium">
                            <optgroup label="Tribunais de Justiça Estaduais">
                                <option value="TJSP" selected>TJSP - Tabela Prática Débitos Judiciais (Lei 14.905)</option>
                                <option value="TJRJ">TJRJ - Tabela Tribunal de Justiça do Rio de Janeiro</option>
                                <option value="TJMG">TJMG - Tabela Tribunal de Justiça de MG (Não Expurgada)</option>
                                <option value="TJMG_EXPURGADA">TJMG - Tabela Tribunal de Justiça de MG (Expurgada)</option>
                                <option value="TJRS">TJRS - Tabela Tribunal de Justiça do Rio Grande do Sul</option>
                                <option value="TJPR">TJPR - Critério Oficial Média (IGP-DI / INPC)</option>
                                <option value="TJPR_PRECAT">TJPR - Precatórios Judiciais (IPCA-E / EC 113)</option>
                                <option value="TJSC">TJSC - Tabela Tribunal de Justiça de Santa Catarina</option>
                                <option value="TJES">TJES - Tabela Tribunal de Justiça do Espírito Santo</option>
                                <option value="TJCE">TJCE - Tabela Tribunal de Justiça do Ceará</option>
                                <option value="TJDF">TJDF - Tabela Tribunal de Justiça do Distrito Federal</option>
                            </optgroup>
                            <optgroup label="Justiça Federal (CJF Res. 267/2013) & Fazenda">
                                <option value="JF_GERAL">Justiça Federal - Ações Condenatórias em Geral (CJF)</option>
                                <option value="JF_PREVIDENCIARIO">Justiça Federal - Benefícios Previdenciários (INPC/IPCA-E)</option>
                                <option value="JF_FAZENDA">Justiça Federal - Dívida da Fazenda Pública (TR / Lei 11.960)</option>
                                <option value="JF_DESAPROPRIACAO">Justiça Federal - Desapropriações (IPCA-E / SELIC)</option>
                            </optgroup>
                            <optgroup label="Justiça do Trabalho (TST / STF)">
                                <option value="TST_TRABALHISTA">TST Trabalhista - Tabela Oficial (ADC 58 STF - IPCA-E / Selic)</option>
                            </optgroup>
                            <optgroup label="Índices Econômicos Oficiais (Bacen / IBGE / FGV)">
                                <option value="IPCA">IPCA - Índice Geral de Preços ao Consumidor Amplo (IBGE)</option>
                                <option value="INPC">INPC - Índice Nacional de Preços ao Consumidor (IBGE)</option>
                                <option value="SELIC">SELIC - Taxa Básica Acumulada Mensal (% a.m.)</option>
                                <option value="IGPM">IGP-M - Índice Geral de Preços do Mercado (FGV)</option>
                                <option value="IGPDI">IGP-DI - Índice Geral de Preços - Disp. Interna (FGV)</option>
                                <option value="INCC">INCC - Índice Nacional de Custo da Construção (FGV)</option>
                                <option value="TR">TR - Taxa Referencial</option>
                                <option value="POUPANCA">Poupança - Rendimento Oficial Acumulado (% a.m.)</option>
                                <option value="TAXA_LEGAL">Taxa Legal de Juros (Lei 14.905/2024 - Bacen)</option>
                            </optgroup>
                        </select>
                    </div>

                    <!-- Data de Atualização -->
                    <div>
                        <label class="block text-xs font-bold text-slate-700 uppercase tracking-wider mb-1.5">
                            Data de Atualização (Data-Base) *
                        </label>
                        <input type="date" id="data_atualizacao" name="data_atualizacao" required value="{{ date('Y-m-d') }}" class="w-full text-xs rounded-xl border-slate-300 focus:border-indigo-500 focus:ring-indigo-500 p-2.5 border">
                    </div>
                </div>

                <!-- Juros de Mora e Fórmulas -->
                <div class="grid grid-cols-1 sm:grid-cols-4 gap-4 pt-4 border-t border-slate-100">
                    <div>
                        <label class="block text-xs font-semibold text-slate-700 mb-1">Regra de Juros de Mora</label>
                        <select id="juros_tipo" name="juros_tipo" class="w-full text-xs rounded-xl border-slate-300 p-2.5 border font-medium">
                            <option value="CONVENCIONAL_MENSAL" selected>Percentual Mensal Convencional</option>
                            <option value="LEI_14905">Nova Taxa Legal (Lei 14.905/2024)</option>
                            <option value="SEM_JUROS">Sem Juros de Mora</option>
                        </select>
                    </div>

                    <div id="boxPercentualJuros">
                        <label class="block text-xs font-semibold text-slate-700 mb-1">Taxa Mensal (% a.m.)</label>
                        <input type="number" step="0.01" id="juros_percentual" name="juros_percentual" value="1.00" class="w-full text-xs rounded-xl border-slate-300 p-2.5 border">
                    </div>

                    <div>
                        <label class="block text-xs font-semibold text-slate-700 mb-1">Regime de Capitalização</label>
                        <select id="juros_modo" name="juros_modo" class="w-full text-xs rounded-xl border-slate-300 p-2.5 border font-medium">
                            <option value="SIMPLES" selected>Juros Simples (Padrão Judicial)</option>
                            <option value="COMPOSTO">Juros Compostos</option>
                        </select>
                    </div>

                    <div class="flex items-center pt-5">
                        <label class="inline-flex items-center gap-2 cursor-pointer text-xs font-medium text-slate-700">
                            <input type="checkbox" id="aplicar_deflacao" name="aplicar_deflacao" class="rounded border-slate-300 text-indigo-600 focus:ring-indigo-500">
                            Aplicar Deflação Negativa
                        </label>
                    </div>
                </div>

                <!-- Cumprimento de Sentença (Art. 523 CPC e Honorários) -->
                <div class="p-4 rounded-xl bg-slate-50 border border-slate-200">
                    <span class="text-xs font-bold text-slate-800 uppercase tracking-wider block mb-2">
                        Fase de Execução / Cumprimento de Sentença (Opcional)
                    </span>
                    <div class="grid grid-cols-1 sm:grid-cols-4 gap-4 text-xs">
                        <label class="inline-flex items-center gap-2 cursor-pointer font-medium text-slate-700">
                            <input type="checkbox" id="aplicar_art_523_multa" name="aplicar_art_523_multa" class="rounded border-slate-300 text-indigo-600 focus:ring-indigo-500">
                            Multa Art. 523, §1º CPC (+10%)
                        </label>

                        <label class="inline-flex items-center gap-2 cursor-pointer font-medium text-slate-700">
                            <input type="checkbox" id="aplicar_art_523_honorarios" name="aplicar_art_523_honorarios" class="rounded border-slate-300 text-indigo-600 focus:ring-indigo-500">
                            Honorários Art. 523, §1º (+10%)
                        </label>

                        <div>
                            <span class="block text-slate-500 mb-1">Multa Moratória Contratual (%)</span>
                            <input type="number" step="0.01" id="multa_moratoria_percentual" value="0.0" class="w-full text-xs rounded-lg border-slate-300 p-1.5 border">
                        </div>

                        <div>
                            <span class="block text-slate-500 mb-1">Honorários Sucumbenciais (%)</span>
                            <input type="number" step="0.01" id="honorarios_sucumbenciais_percentual" value="0.0" class="w-full text-xs rounded-lg border-slate-300 p-1.5 border">
                        </div>
                    </div>
                </div>
            </div>

            <!-- Card 2: Lançamentos e Importador de Planilha -->
            <div class="bg-white rounded-2xl shadow-xl border border-slate-200 p-6 sm:p-8 space-y-5">
                <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-100 pb-4">
                    <div>
                        <h2 class="text-base font-bold text-slate-900 flex items-center gap-2">
                            <span class="w-6 h-6 rounded-lg bg-indigo-100 text-indigo-700 flex items-center justify-center text-xs font-black">2</span>
                            Tabela de Parcelas, Custas e Abatimentos
                        </h2>
                        <span class="text-xs text-slate-400">Lance as parcelas manuais ou importe dezenas via planilha Excel / CSV</span>
                    </div>

                    <!-- Botões de Ação da Planilha -->
                    <div class="flex items-center gap-2">
                        <!-- Botão Importar Arquivo -->
                        <label class="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 bg-emerald-50 hover:bg-emerald-100 text-emerald-800 border border-emerald-300 rounded-xl text-xs font-bold transition-all shadow-xs">
                            <svg class="w-4 h-4 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"/>
                            </svg>
                            Importar Planilha (.xlsx, .csv)
                            <input type="file" id="arquivoPlanilhaInput" accept=".xlsx,.csv,.tsv,.txt" class="hidden" onchange="uploadPlanilha(this)">
                        </label>

                        <!-- Botão Colar do Excel -->
                        <button type="button" onclick="abrirModalColar()" class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-800 border border-slate-300 rounded-xl text-xs font-bold transition-all">
                            <svg class="w-4 h-4 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
                            </svg>
                            Colar do Excel
                        </button>
                    </div>
                </div>

                <!-- Tabela de Parcelas Dinâmica -->
                <div class="overflow-x-auto border border-slate-200 rounded-xl">
                    <table class="w-full text-left text-xs border-collapse" id="tabelaParcelas">
                        <thead>
                            <tr class="bg-slate-100 text-slate-700 font-bold border-b border-slate-200">
                                <th class="p-2.5 w-10 text-center">#</th>
                                <th class="p-2.5">Descrição</th>
                                <th class="p-2.5 w-44">Natureza</th>
                                <th class="p-2.5 w-36">Vencimento</th>
                                <th class="p-2.5 w-40">Valor Nominal (R$)</th>
                                <th class="p-2.5 w-16 text-center">Excluir</th>
                            </tr>
                        </thead>
                        <tbody id="corpoParcelas" class="divide-y divide-slate-100">
                            <!-- Linha Padrão Inicial -->
                            <tr class="item-parcela hover:bg-slate-50">
                                <td class="p-2.5 text-center text-slate-400 font-bold index-num">1</td>
                                <td class="p-2.5">
                                    <input type="text" class="campo-desc w-full text-xs rounded-lg border-slate-200 p-1.5 border" value="Parcela 01 - Principal">
                                </td>
                                <td class="p-2.5">
                                    <select class="campo-tipo w-full text-xs rounded-lg border-slate-200 p-1.5 border font-medium">
                                        <option value="DEBITO" selected>Débito Principal</option>
                                        <option value="CUSTA_PROCESSUAL">Custa Processual</option>
                                        <option value="DESPESA_PROCESSUAL">Despesa</option>
                                        <option value="PAGAMENTO_PARCIAL">Abatimento</option>
                                    </select>
                                </td>
                                <td class="p-2.5">
                                    <input type="date" class="campo-data w-full text-xs rounded-lg border-slate-200 p-1.5 border" value="{{ date('Y-m-d', strtotime('-2 years')) }}">
                                </td>
                                <td class="p-2.5">
                                    <input type="number" step="0.01" class="campo-valor w-full text-xs rounded-lg border-slate-200 p-1.5 border font-mono font-bold" value="10000.00">
                                </td>
                                <td class="p-2.5 text-center">
                                    <button type="button" onclick="removerLinha(this)" class="text-slate-400 hover:text-rose-600 transition-colors p-1">
                                        &times;
                                    </button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>

                <div class="flex items-center justify-between">
                    <button type="button" onclick="adicionarLinha()" class="inline-flex items-center gap-1.5 px-4 py-2 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 text-xs font-bold rounded-xl transition-all">
                        + Adicionar Parcela
                    </button>
                    <span id="contadorParcelas" class="text-xs text-slate-500">1 parcela cadastrada</span>
                </div>
            </div>

            <!-- Barra de Ação Principal -->
            <div class="flex flex-col sm:flex-row items-center justify-between gap-4 bg-white p-5 rounded-2xl shadow-md border border-slate-200">
                <div class="text-xs text-slate-500">
                    Clique abaixo para executar a liquidação contábil com precisão matemática em Go.
                </div>
                <div class="flex items-center gap-3 w-full sm:w-auto">
                    <button type="button" onclick="limparFormulario()" class="w-full sm:w-auto px-5 py-2.5 rounded-xl border border-slate-300 text-xs font-semibold text-slate-700 hover:bg-slate-100 transition-all">
                        Limpar
                    </button>
                    <button type="submit" id="btnCalcular" class="w-full sm:w-auto px-8 py-3 rounded-xl bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-extrabold shadow-lg shadow-indigo-600/30 transition-all flex items-center justify-center gap-2">
                        <span id="btnCalcularTexto">CALCULAR DÉBITO JUDICIAL</span>
                        <svg id="btnCalcularLoading" class="animate-spin -mr-1 ml-2 h-4 w-4 text-white hidden" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
                        </svg>
                    </button>
                </div>
            </div>
        </form>

        <!-- Seção de Resultados (Exibida dinamicamente após o cálculo) -->
        <div id="secaoResultados" class="hidden space-y-6">
            
            <!-- 4 Cards de Totais -->
            <div class="grid grid-cols-1 sm:grid-cols-4 gap-4">
                <div class="bg-white p-5 rounded-2xl shadow-md border border-slate-200">
                    <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider block">Principal Corrigido</span>
                    <div id="resPrincipalCorrigido" class="text-xl font-bold text-slate-900 mt-1 font-mono-code">R$ 0,00</div>
                    <span class="text-[10px] text-slate-400">Valor nominal + inflação acumulada</span>
                </div>

                <div class="bg-white p-5 rounded-2xl shadow-md border border-slate-200">
                    <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider block">Juros Moratórios</span>
                    <div id="resJurosTotal" class="text-xl font-bold text-indigo-600 mt-1 font-mono-code">R$ 0,00</div>
                    <span class="text-[10px] text-slate-400">Juros de mora acumulados</span>
                </div>

                <div class="bg-white p-5 rounded-2xl shadow-md border border-slate-200">
                    <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider block">Multas e Honorários</span>
                    <div id="resMultasTotal" class="text-xl font-bold text-amber-600 mt-1 font-mono-code">R$ 0,00</div>
                    <span class="text-[10px] text-slate-400">Encargos de execução</span>
                </div>

                <div class="bg-gradient-to-br from-indigo-900 to-slate-900 p-5 rounded-2xl shadow-lg text-white border border-indigo-950">
                    <span class="text-xs font-bold text-indigo-300 uppercase tracking-wider block">VALOR TOTAL ATUALIZADO</span>
                    <div id="resTotalGeral" class="text-2xl font-black text-emerald-400 mt-1 font-mono-code">R$ 0,00</div>
                    <span class="text-[10px] text-slate-400">Total geral liquidado</span>
                </div>
            </div>

            <!-- Banner Lead Magnet: Emissão de Laudo Certificado -->
            <div class="bg-gradient-to-r from-emerald-600 via-teal-600 to-indigo-700 rounded-2xl p-6 sm:p-8 text-white shadow-xl flex flex-col sm:flex-row items-center justify-between gap-6">
                <div class="space-y-1 text-center sm:text-left">
                    <span class="inline-flex items-center gap-1.5 px-3 py-0.5 rounded-full text-xs font-bold bg-white/20 text-white uppercase tracking-wider">
                        Documento Oficial para Petição
                    </span>
                    <h3 class="text-2xl font-black tracking-tight">Emitir Laudo Pericial Certificado (PDF)</h3>
                    <p class="text-emerald-100 text-xs max-w-xl">
                        Gere a memória de cálculo formal em formato A4 para juntada processual, acompanhada de <strong>QR Code e Hash SHA-256 de autenticidade pública</strong> que comprova que o cálculo não foi adulterado.
                    </p>
                </div>
                <button type="button" onclick="abrirModalLead()" class="shrink-0 px-8 py-3.5 bg-white hover:bg-slate-100 text-slate-900 text-sm font-black rounded-xl shadow-lg hover:shadow-xl transition-all">
                    BAIXAR LAUDO COM QR CODE &rarr;
                </button>
            </div>

            <!-- Tabela Analítica de Itens Liquidados -->
            <div class="bg-white rounded-2xl shadow-md border border-slate-200 p-6 space-y-4">
                <h4 class="text-sm font-bold text-slate-900 uppercase tracking-wider flex items-center gap-2">
                    <span class="w-1.5 h-3 bg-indigo-600 rounded-sm"></span>
                    Demonstrativo Discriminado das Parcelas
                </h4>
                <div class="overflow-x-auto border border-slate-200 rounded-xl">
                    <table class="w-full text-left text-xs border-collapse" id="tabelaResultadoItens">
                        <thead>
                            <tr class="bg-slate-100 text-slate-700 font-bold border-b border-slate-200">
                                <th class="p-2.5 text-center">#</th>
                                <th class="p-2.5">Descrição</th>
                                <th class="p-2.5">Vencimento</th>
                                <th class="p-2.5 text-right">Valor Nominal</th>
                                <th class="p-2.5 text-right">Fator Corr.</th>
                                <th class="p-2.5 text-right">Principal Corr.</th>
                                <th class="p-2.5 text-right">Juros</th>
                                <th class="p-2.5 text-right">Multa</th>
                                <th class="p-2.5 text-right font-black">Subtotal (R$)</th>
                            </tr>
                        </thead>
                        <tbody id="corpoResultadoItens" class="divide-y divide-slate-100 font-mono-code text-[11px]">
                            <!-- Preenchido via JS -->
                        </tbody>
                    </table>
                </div>

                <!-- Notas Legais -->
                <div id="boxNotasLegais" class="p-4 rounded-xl bg-slate-50 border border-slate-200 text-xs text-slate-600 space-y-1">
                    <span class="font-bold text-slate-800 uppercase text-[10px] block">Notas Técnicas:</span>
                    <ul id="listaNotasLegais" class="list-disc list-inside space-y-1"></ul>
                </div>
            </div>

        </div>

    </main>

    <!-- Modal: Colar do Excel -->
    <div id="modalColar" class="fixed inset-0 z-50 bg-slate-900/60 backdrop-blur-xs flex items-center justify-center p-4 hidden">
        <div class="bg-white rounded-2xl max-w-xl w-full p-6 shadow-2xl space-y-4 border border-slate-200">
            <div class="flex items-center justify-between border-b border-slate-100 pb-3">
                <h3 class="text-base font-bold text-slate-900">Colar Parcelas do Excel / Planilha</h3>
                <button type="button" onclick="fecharModalColar()" class="text-slate-400 hover:text-slate-600 text-xl font-bold">&times;</button>
            </div>
            <p class="text-xs text-slate-500">
                Copie as linhas no seu Excel (Ctrl+C) com colunas contendo data, descrição e valor, e cole abaixo:
            </p>
            <textarea id="textoColado" rows="8" placeholder="Exemplo:&#10;15/01/2023	Parcela 01	1500,00&#10;15/02/2023	Parcela 02	1500,00" class="w-full text-xs font-mono rounded-xl border border-slate-300 p-3 focus:ring-indigo-500 focus:border-indigo-500"></textarea>
            <div class="flex items-center justify-end gap-3 pt-2">
                <button type="button" onclick="fecharModalColar()" class="px-4 py-2 rounded-xl text-xs font-semibold text-slate-600 hover:bg-slate-100">
                    Cancelar
                </button>
                <button type="button" onclick="processarTextoColado()" class="px-5 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl text-xs font-bold shadow-md">
                    Processar e Adicionar
                </button>
            </div>
        </div>
    </div>

    <!-- Modal: Captura de Lead & Emissão do Laudo Certificado -->
    <div id="modalLead" class="fixed inset-0 z-50 bg-slate-900/70 backdrop-blur-xs flex items-center justify-center p-4 hidden">
        <div class="bg-white rounded-2xl max-w-lg w-full p-6 sm:p-8 shadow-2xl space-y-5 border border-slate-200">
            <div class="flex items-start justify-between border-b border-slate-100 pb-3">
                <div class="space-y-0.5">
                    <span class="text-[10px] font-bold text-emerald-600 uppercase tracking-wider block">Certificação Digital</span>
                    <h3 class="text-lg font-black text-slate-900">Emissão de Laudo Pericial Certificado</h3>
                </div>
                <button type="button" onclick="fecharModalLead()" class="text-slate-400 hover:text-slate-600 text-xl font-bold">&times;</button>
            </div>

            <p class="text-xs text-slate-500">
                Informe os dados do responsável pelo cálculo para constar no cabeçalho do Laudo Oficial e gerar o QR Code de autenticidade:
            </p>

            <form id="formLead" onsubmit="event.preventDefault(); submeterLead();" class="space-y-4">
                <div>
                    <label class="block text-xs font-bold text-slate-700 uppercase mb-1">Nome Completo / Perito / Advogado *</label>
                    <input type="text" id="lead_nome" required placeholder="Ex: Dr. Carlos Eduardo Silva" class="w-full text-xs rounded-xl border border-slate-300 p-2.5">
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div>
                        <label class="block text-xs font-bold text-slate-700 uppercase mb-1">E-mail Profissional *</label>
                        <input type="email" id="lead_email" required placeholder="carlos@escritorio.adv.br" class="w-full text-xs rounded-xl border border-slate-300 p-2.5">
                    </div>
                    <div>
                        <label class="block text-xs font-bold text-slate-700 uppercase mb-1">WhatsApp (com DDD) *</label>
                        <input type="text" id="lead_whatsapp" required placeholder="(11) 98765-4321" class="w-full text-xs rounded-xl border border-slate-300 p-2.5">
                    </div>
                </div>

                <div>
                    <label class="block text-xs font-bold text-slate-700 uppercase mb-1">Inscrição OAB / UF ou Órgão Pericial</label>
                    <input type="text" id="lead_oab" placeholder="Ex: OAB/SP 123.456 ou CONPEJ" class="w-full text-xs rounded-xl border border-slate-300 p-2.5">
                </div>

                <div class="pt-2">
                    <button type="submit" id="btnEmitirLaudo" class="w-full py-3.5 bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-black rounded-xl shadow-lg transition-all flex items-center justify-center gap-2">
                        <span id="btnEmitirLaudoTexto">GERAR LAUDO E VISUALIZAR PDF COM QR CODE</span>
                        <svg id="btnEmitirLaudoLoading" class="animate-spin -mr-1 ml-2 h-4 w-4 text-white hidden" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
                        </svg>
                    </button>
                </div>
            </form>
        </div>
    </div>

    <!-- Rodapé -->
    <footer class="bg-white border-t border-slate-200 py-6 text-center text-xs text-slate-500">
        <p>&copy; {{ date('Y') }} Robô Jurídico Inteligente &bull; Motor de Liquidação Judicial em Go com Paridade DrCalc.</p>
    </footer>

    <!-- Scripts da Calculadora Interativa -->
    <script>
        let ultimoCalculoPayload = null;
        let ultimoCalculoResultado = null;

        // Toggle do campo de juros
        document.getElementById('juros_tipo').addEventListener('change', function() {
            const box = document.getElementById('boxPercentualJuros');
            box.style.display = (this.value === 'CONVENCIONAL_MENSAL') ? 'block' : 'none';
        });

        function adicionarLinha(dados = {}) {
            const corpo = document.getElementById('corpoParcelas');
            const totalLinhas = corpo.querySelectorAll('.item-parcela').length + 1;

            const tr = document.createElement('tr');
            tr.className = 'item-parcela hover:bg-slate-50';
            tr.innerHTML = `
                <td class="p-2.5 text-center text-slate-400 font-bold index-num">${totalLinhas}</td>
                <td class="p-2.5">
                    <input type="text" class="campo-desc w-full text-xs rounded-lg border-slate-200 p-1.5 border" value="${dados.descricao || 'Parcela ' + totalLinhas}">
                </td>
                <td class="p-2.5">
                    <select class="campo-tipo w-full text-xs rounded-lg border-slate-200 p-1.5 border font-medium">
                        <option value="DEBITO" ${dados.tipo === 'DEBITO' || !dados.tipo ? 'selected' : ''}>Débito Principal</option>
                        <option value="CUSTA_PROCESSUAL" ${dados.tipo === 'CUSTA_PROCESSUAL' ? 'selected' : ''}>Custa Processual</option>
                        <option value="DESPESA_PROCESSUAL" ${dados.tipo === 'DESPESA_PROCESSUAL' ? 'selected' : ''}>Despesa</option>
                        <option value="PAGAMENTO_PARCIAL" ${dados.tipo === 'PAGAMENTO_PARCIAL' ? 'selected' : ''}>Abatimento</option>
                    </select>
                </td>
                <td class="p-2.5">
                    <input type="date" class="campo-data w-full text-xs rounded-lg border-slate-200 p-1.5 border" value="${dados.data || '{{ date('Y-m-d') }}'}">
                </td>
                <td class="p-2.5">
                    <input type="number" step="0.01" class="campo-valor w-full text-xs rounded-lg border-slate-200 p-1.5 border font-mono font-bold" value="${dados.valor || '1000.00'}">
                </td>
                <td class="p-2.5 text-center">
                    <button type="button" onclick="removerLinha(this)" class="text-slate-400 hover:text-rose-600 transition-colors p-1">
                        &times;
                    </button>
                </td>
            `;
            corpo.appendChild(tr);
            atualizarContador();
        }

        function removerLinha(btn) {
            const corpo = document.getElementById('corpoParcelas');
            if (corpo.querySelectorAll('.item-parcela').length <= 1) {
                alert('É necessário manter pelo menos uma parcela para liquidação.');
                return;
            }
            btn.closest('tr').remove();
            atualizarContador();
        }

        function atualizarContador() {
            const linhas = document.querySelectorAll('.item-parcela');
            linhas.forEach((tr, idx) => {
                const numEl = tr.querySelector('.index-num');
                if (numEl) numEl.textContent = idx + 1;
            });
            document.getElementById('contadorParcelas').textContent = `${linhas.length} parcela(s) cadastrada(s)`;
        }

        function formatarMoeda(val) {
            return Number(val || 0).toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
        }

        async function executarCalculo() {
            const btn = document.getElementById('btnCalcular');
            const btnTexto = document.getElementById('btnCalcularTexto');
            const btnLoading = document.getElementById('btnCalcularLoading');

            btn.disabled = true;
            btnLoading.classList.remove('hidden');
            btnTexto.textContent = 'PROCESSANDO...';

            // Coleta dados das parcelas
            const linhas = document.querySelectorAll('.item-parcela');
            const parcelas = [];
            linhas.forEach((tr, idx) => {
                parcelas.push({
                    numero_item: idx + 1,
                    descricao: tr.querySelector('.campo-desc').value || `Parcela ${idx + 1}`,
                    tipo: tr.querySelector('.campo-tipo').value || 'DEBITO',
                    data_vencimento: tr.querySelector('.campo-data').value,
                    valor_original: parseFloat(tr.querySelector('.campo-valor').value || 0)
                });
            });

            const payload = {
                numero_processo: document.getElementById('numero_processo').value,
                indexador_codigo: document.getElementById('indexador_codigo').value,
                data_atualizacao: document.getElementById('data_atualizacao').value,
                aplicar_deflacao: document.getElementById('aplicar_deflacao').checked,
                juros_mora: {
                    tipo: document.getElementById('juros_tipo').value,
                    percentual_mensal: parseFloat(document.getElementById('juros_percentual').value || 1.0),
                    modo: document.getElementById('juros_modo').value,
                    aplicar_prorata_die: true
                },
                fase_execucao: {
                    aplicar_art_523_multa: document.getElementById('aplicar_art_523_multa').checked,
                    aplicar_art_523_honorarios: document.getElementById('aplicar_art_523_honorarios').checked,
                    multa_moratoria_percentual: parseFloat(document.getElementById('multa_moratoria_percentual').value || 0),
                    honorarios_sucumbenciais_percentual: parseFloat(document.getElementById('honorarios_sucumbenciais_percentual').value || 0)
                },
                parcelas: parcelas
            };

            ultimoCalculoPayload = payload;

            try {
                const response = await fetch('{{ route('calculadora.calcular') }}', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-TOKEN': '{{ csrf_token() }}'
                    },
                    body: JSON.stringify(payload)
                });

                const res = await response.json();

                if (!res.sucesso && !res.ok && res.mensagem) {
                    alert('Erro no cálculo: ' + (res.mensagem || 'Falha ao processar'));
                    return;
                }

                ultimoCalculoResultado = res;
                exibirResultados(res);
            } catch (err) {
                alert('Erro de comunicação com o servidor: ' + err.message);
            } finally {
                btn.disabled = false;
                btnLoading.classList.add('hidden');
                btnTexto.textContent = 'CALCULAR DÉBITO JUDICIAL';
            }
        }

        function exibirResultados(res) {
            document.getElementById('secaoResultados').classList.remove('hidden');

            const resumo = res.resumo || {};
            document.getElementById('resPrincipalCorrigido').textContent = formatarMoeda(resumo.total_principal_corrigido);
            document.getElementById('resJurosTotal').textContent = formatarMoeda(resumo.total_juros);
            document.getElementById('resMultasTotal').textContent = formatarMoeda((resumo.total_multas || 0) + (resumo.total_honorarios || 0));
            document.getElementById('resTotalGeral').textContent = formatarMoeda(resumo.total_geral);

            // Popula tabela de itens
            const corpo = document.getElementById('corpoResultadoItens');
            corpo.innerHTML = '';
            (res.itens || []).forEach(it => {
                const tr = document.createElement('tr');
                tr.className = 'hover:bg-slate-50';
                tr.innerHTML = `
                    <td class="p-2.5 text-center text-slate-400">${it.numero_item}</td>
                    <td class="p-2.5 font-sans font-medium text-slate-800">${it.descricao}</td>
                    <td class="p-2.5 text-slate-600">${it.data_vencimento}</td>
                    <td class="p-2.5 text-right">${formatarMoeda(it.valor_original)}</td>
                    <td class="p-2.5 text-right text-slate-500">${Number(it.fator_correcao).toFixed(6)}</td>
                    <td class="p-2.5 text-right font-semibold text-slate-900">${formatarMoeda(it.valor_corrigido)}</td>
                    <td class="p-2.5 text-right text-indigo-700 font-semibold">${formatarMoeda(it.valor_juros)}</td>
                    <td class="p-2.5 text-right text-amber-700">${formatarMoeda(it.valor_multa_moratoria)}</td>
                    <td class="p-2.5 text-right font-black text-slate-950 bg-slate-50">${formatarMoeda(it.subtotal)}</td>
                `;
                corpo.appendChild(tr);
            });

            // Popula notas legais
            const listaNotas = document.getElementById('listaNotasLegais');
            listaNotas.innerHTML = '';
            (res.notas_legais || []).forEach(nota => {
                const li = document.createElement('li');
                li.textContent = nota;
                listaNotas.appendChild(li);
            });

            // Rola suavemente até os resultados
            document.getElementById('secaoResultados').scrollIntoView({ behavior: 'smooth' });
        }

        // Upload de Planilha (.xlsx, .csv)
        async function uploadPlanilha(input) {
            if (!input.files || input.files.length === 0) return;

            const formData = new FormData();
            formData.append('planilha', input.files[0]);

            try {
                const res = await fetch('{{ route('calculadora.importar-planilha') }}', {
                    method: 'POST',
                    headers: { 'X-CSRF-TOKEN': '{{ csrf_token() }}' },
                    body: formData
                });
                const data = await res.json();
                if (data.ok && data.parcelas && data.parcelas.length > 0) {
                    carregarParcelasImportadas(data.parcelas);
                    alert(`Sucesso! ${data.parcelas.length} parcela(s) importada(s) do arquivo.`);
                } else {
                    alert('Falha ao importar: ' + (data.mensagem || 'Formato não reconhecido'));
                }
            } catch (e) {
                alert('Erro ao enviar planilha: ' + e.message);
            } finally {
                input.value = '';
            }
        }

        // Colar do Excel
        function abrirModalColar() {
            document.getElementById('modalColar').classList.remove('hidden');
        }
        function fecharModalColar() {
            document.getElementById('modalColar').classList.add('hidden');
            document.getElementById('textoColado').value = '';
        }

        async function processarTextoColado() {
            const texto = document.getElementById('textoColado').value;
            if (!texto.trim()) {
                alert('Cole o conteúdo das linhas antes de processar.');
                return;
            }

            try {
                const res = await fetch('{{ route('calculadora.importar-planilha') }}', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-TOKEN': '{{ csrf_token() }}'
                    },
                    body: JSON.stringify({ texto: texto })
                });
                const data = await res.json();
                if (data.ok && data.parcelas && data.parcelas.length > 0) {
                    carregarParcelasImportadas(data.parcelas);
                    fecharModalColar();
                    alert(`${data.parcelas.length} parcela(s) adicionada(s) com sucesso!`);
                } else {
                    alert('Falha ao interpretar as linhas: ' + (data.mensagem || 'Verifique o formato das colunas.'));
                }
            } catch (e) {
                alert('Erro ao processar texto: ' + e.message);
            }
        }

        function carregarParcelasImportadas(parcelas) {
            const corpo = document.getElementById('corpoParcelas');
            corpo.innerHTML = '';
            parcelas.forEach(p => {
                adicionarLinha({
                    descricao: p.descricao,
                    data: p.data,
                    valor: p.valor_original,
                    tipo: 'DEBITO'
                });
            });
            atualizarContador();
        }

        // Modal de Lead e Emissão de Laudo
        function abrirModalLead() {
            if (!ultimoCalculoResultado) {
                alert('Realize o cálculo primeiro antes de emitir o laudo.');
                return;
            }
            document.getElementById('modalLead').classList.remove('hidden');
        }
        function fecharModalLead() {
            document.getElementById('modalLead').classList.add('hidden');
        }

        async function submeterLead() {
            const btn = document.getElementById('btnEmitirLaudo');
            const btnTexto = document.getElementById('btnEmitirLaudoTexto');
            const btnLoading = document.getElementById('btnEmitirLaudoLoading');

            btn.disabled = true;
            btnLoading.classList.remove('hidden');
            btnTexto.textContent = 'GERANDO CERTIFICADO...';

            const payloadLead = {
                nome: document.getElementById('lead_nome').value,
                email: document.getElementById('lead_email').value,
                whatsapp: document.getElementById('lead_whatsapp').value,
                oab: document.getElementById('lead_oab').value,
                dados_calculo: ultimoCalculoPayload,
                resultado: ultimoCalculoResultado
            };

            try {
                const res = await fetch('{{ route('calculadora.emitir-laudo') }}', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-TOKEN': '{{ csrf_token() }}'
                    },
                    body: JSON.stringify(payloadLead)
                });
                const data = await res.json();
                if (data.ok && data.url_laudo) {
                    fecharModalLead();
                    window.open(data.url_laudo, '_blank');
                } else {
                    alert('Erro ao emitir laudo: ' + (data.message || 'Verifique os dados informados.'));
                }
            } catch (e) {
                alert('Erro na emissão: ' + e.message);
            } finally {
                btn.disabled = false;
                btnLoading.classList.add('hidden');
                btnTexto.textContent = 'GERAR LAUDO E VISUALIZAR PDF COM QR CODE';
            }
        }

        function limparFormulario() {
            document.getElementById('formCalculo').reset();
            document.getElementById('secaoResultados').classList.add('hidden');
            ultimoCalculoPayload = null;
            ultimoCalculoResultado = null;
        }
    </script>
</body>
</html>
