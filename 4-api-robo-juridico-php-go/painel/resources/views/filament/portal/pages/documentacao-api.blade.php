<x-filament-panels::page>
    <div class="space-y-6">
        <!-- Card de Boas-Vindas e Autenticação -->
        <x-filament::section>
            <x-slot name="heading">
                Autenticação e Acesso à API
            </x-slot>
            <x-slot name="description">
                Padrões de cabeçalho HTTP e credenciais para integração
            </x-slot>

            <div class="space-y-4">
                <p class="text-sm text-gray-700 dark:text-gray-300 leading-relaxed">
                    Todas as chamadas à API devem incluir o seu token nos cabeçalhos HTTP. A API aceita nativamente tanto
                    <code class="px-1.5 py-0.5 rounded text-xs font-mono font-semibold bg-gray-100 text-gray-900 border border-gray-300 dark:bg-gray-800 dark:text-gray-200 dark:border-gray-700">Authorization: Bearer &lt;seu_token&gt;</code>
                    quanto
                    <code class="px-1.5 py-0.5 rounded text-xs font-mono font-semibold bg-gray-100 text-gray-900 border border-gray-300 dark:bg-gray-800 dark:text-gray-200 dark:border-gray-700">X-API-Key: &lt;seu_token&gt;</code>.
                    Você pode gerar e revogar seus tokens a qualquer momento no menu <strong>Minhas Chaves de API</strong>.
                </p>

                <div class="p-4 rounded-xl font-mono text-xs overflow-x-auto shadow-inner" style="background-color: #0f172a; color: #f8fafc; border: 1px solid #1e293b;">
                    <div class="mb-1" style="color: #94a3b8;"># URL Base da API:</div>
                    <div class="flex items-center gap-2 flex-wrap">
                        <strong class="text-sm font-bold" style="color: #34d399;">http://localhost:8000</strong>
                        <span style="color: #64748b;">|</span>
                        <span class="text-xs" style="color: #cbd5e1;">Prefixos suportados: <code class="px-1.5 py-0.5 rounded" style="background-color: #1e293b; color: #38bdf8;">/v1</code> e <code class="px-1.5 py-0.5 rounded" style="background-color: #1e293b; color: #38bdf8;">/api/v1</code></span>
                    </div>
                </div>
            </div>
        </x-filament::section>

        <!-- Exemplo cURL GET -->
        <x-filament::section>
            <x-slot name="heading">
                1. Consulta Canônica por Rota REST (cURL GET — Recomendado)
            </x-slot>
            <x-slot name="description">
                Método padrão para consulta unificada de processo via URL canônica
            </x-slot>

            <div class="p-4 rounded-xl font-mono text-xs overflow-x-auto shadow-inner" style="background-color: #0f172a; color: #f8fafc; border: 1px solid #1e293b;">
                <pre class="m-0 p-0 font-mono"><code><span style="color: #f43f5e; font-weight: bold;">curl</span> -X GET http://localhost:8000/v1/processos/5053283-30.2024.8.13.0079 \
  -H <span style="color: #38bdf8;">"Authorization: Bearer rj_live_SEU_TOKEN_AQUI"</span></code></pre>
            </div>
        </x-filament::section>

        <!-- Exemplo cURL POST -->
        <x-filament::section>
            <x-slot name="heading">
                2. Consulta via Corpo JSON (cURL POST)
            </x-slot>
            <x-slot name="description">
                Útil quando o número do processo é enviado no payload de requisições POST
            </x-slot>

            <div class="p-4 rounded-xl font-mono text-xs overflow-x-auto shadow-inner" style="background-color: #0f172a; color: #f8fafc; border: 1px solid #1e293b;">
                <pre class="m-0 p-0 font-mono"><code><span style="color: #f43f5e; font-weight: bold;">curl</span> -X POST http://localhost:8000/api/v1/processos/consultar \
  -H <span style="color: #38bdf8;">"Authorization: Bearer rj_live_SEU_TOKEN_AQUI"</span> \
  -H <span style="color: #38bdf8;">"Content-Type: application/json"</span> \
  -d <span style="color: #a7f3d0;">'{"numero_processo": "5053283-30.2024.8.13.0079"}'</span></code></pre>
            </div>
        </x-filament::section>

        <!-- Exemplo Python -->
        <x-filament::section>
            <x-slot name="heading">
                3. Integração em Python (requests)
            </x-slot>
            <x-slot name="description">
                Exemplo pronto para scripts, bots e serviços em Python
            </x-slot>

            <div class="p-4 rounded-xl font-mono text-xs overflow-x-auto shadow-inner" style="background-color: #0f172a; color: #f8fafc; border: 1px solid #1e293b;">
                <pre class="m-0 p-0 font-mono"><code><span style="color: #c084fc; font-weight: bold;">import</span> requests

url = <span style="color: #38bdf8;">"http://localhost:8000/v1/processos/5053283-30.2024.8.13.0079"</span>
headers = {
    <span style="color: #38bdf8;">"Authorization"</span>: <span style="color: #38bdf8;">"Bearer rj_live_SEU_TOKEN_AQUI"</span>
}

response = requests.get(url, headers=headers)
<span style="color: #c084fc; font-weight: bold;">print</span>(<span style="color: #38bdf8;">"Status:"</span>, response.status_code)
<span style="color: #c084fc; font-weight: bold;">print</span>(<span style="color: #38bdf8;">"Dados:"</span>, response.json())</code></pre>
            </div>
        </x-filament::section>

        <!-- Exemplo Node.js -->
        <x-filament::section>
            <x-slot name="heading">
                4. Integração em Node.js / TypeScript (fetch)
            </x-slot>
            <x-slot name="description">
                Compatível nativamente com Node 18+, Bun, Deno e navegadores
            </x-slot>

            <div class="p-4 rounded-xl font-mono text-xs overflow-x-auto shadow-inner" style="background-color: #0f172a; color: #f8fafc; border: 1px solid #1e293b;">
                <pre class="m-0 p-0 font-mono"><code><span style="color: #c084fc; font-weight: bold;">const</span> response = <span style="color: #c084fc; font-weight: bold;">await</span> fetch(<span style="color: #38bdf8;">'http://localhost:8000/v1/processos/5053283-30.2024.8.13.0079'</span>, {
  method: <span style="color: #38bdf8;">'GET'</span>,
  headers: {
    <span style="color: #38bdf8;">'Authorization'</span>: <span style="color: #38bdf8;">'Bearer rj_live_SEU_TOKEN_AQUI'</span>
  }
});

<span style="color: #c084fc; font-weight: bold;">const</span> data = <span style="color: #c084fc; font-weight: bold;">await</span> response.json();
console.log(data);</code></pre>
            </div>
        </x-filament::section>

        <!-- Tabela de Status Codes -->
        <x-filament::section>
            <x-slot name="heading">
                Principais Respostas HTTP
            </x-slot>
            <x-slot name="description">
                Tabela de códigos de resposta, significados e impacto na franquia mensal
            </x-slot>

            <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-800">
                <table class="w-full text-left text-sm">
                    <thead class="bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200 text-xs font-semibold uppercase tracking-wider">
                        <tr>
                            <th class="p-3.5">Código</th>
                            <th class="p-3.5">Significado</th>
                            <th class="p-3.5">Cobrável na Franquia?</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-200 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
                        <tr class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition">
                            <td class="p-3.5 whitespace-nowrap">
                                <x-filament::badge color="success">
                                    200 OK
                                </x-filament::badge>
                            </td>
                            <td class="p-3.5 font-medium text-gray-900 dark:text-gray-100">
                                Processo localizado e retornado com sucesso.
                            </td>
                            <td class="p-3.5 whitespace-nowrap">
                                <x-filament::badge color="success">
                                    Sim
                                </x-filament::badge>
                            </td>
                        </tr>
                        <tr class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition">
                            <td class="p-3.5 whitespace-nowrap">
                                <x-filament::badge color="warning">
                                    400 Bad Request
                                </x-filament::badge>
                            </td>
                            <td class="p-3.5 text-gray-800 dark:text-gray-200">
                                Número de processo com formato inválido (deve seguir CNJ NNNNNNN-DD.AAAA.J.TR.OOOO).
                            </td>
                            <td class="p-3.5 whitespace-nowrap">
                                <x-filament::badge color="gray">
                                    Não
                                </x-filament::badge>
                            </td>
                        </tr>
                        <tr class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition">
                            <td class="p-3.5 whitespace-nowrap">
                                <x-filament::badge color="danger">
                                    401 Unauthorized
                                </x-filament::badge>
                            </td>
                            <td class="p-3.5 text-gray-800 dark:text-gray-200">
                                Token ausente, inválido ou revogado.
                            </td>
                            <td class="p-3.5 whitespace-nowrap">
                                <x-filament::badge color="gray">
                                    Não
                                </x-filament::badge>
                            </td>
                        </tr>
                        <tr class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition">
                            <td class="p-3.5 whitespace-nowrap">
                                <x-filament::badge color="warning">
                                    402 Payment Required
                                </x-filament::badge>
                            </td>
                            <td class="p-3.5 text-gray-800 dark:text-gray-200">
                                Franquia mensal de consultas esgotada. Faça upgrade do plano.
                            </td>
                            <td class="p-3.5 whitespace-nowrap">
                                <x-filament::badge color="gray">
                                    Não
                                </x-filament::badge>
                            </td>
                        </tr>
                        <tr class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition">
                            <td class="p-3.5 whitespace-nowrap">
                                <x-filament::badge color="warning">
                                    429 Too Many Requests
                                </x-filament::badge>
                            </td>
                            <td class="p-3.5 text-gray-800 dark:text-gray-200">
                                Limite de requisições por minuto excedido (Rate limit ativo).
                            </td>
                            <td class="p-3.5 whitespace-nowrap">
                                <x-filament::badge color="gray">
                                    Não
                                </x-filament::badge>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </x-filament::section>
    </div>
</x-filament-panels::page>
