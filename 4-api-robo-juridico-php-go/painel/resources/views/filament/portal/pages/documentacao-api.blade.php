<x-filament-panels::page>
    <div class="space-y-6">
        <!-- Card de Boas-Vindas e Autenticação -->
        <div class="p-6 bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700">
            <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-2">Autenticação e Acesso</h3>
            <p class="text-sm text-gray-600 dark:text-gray-300 mb-4">
                Todas as chamadas à API devem incluir o seu token nos cabeçalhos HTTP via <code>Authorization: Bearer &lt;seu_token&gt;</code>.
                Você pode gerar e revogar seus tokens a qualquer momento no menu <strong>Minhas Chaves de API</strong>.
            </p>
            <div class="bg-gray-900 text-gray-100 p-4 rounded-lg font-mono text-xs overflow-x-auto">
                <span class="text-gray-400"># URL Base da API em produção ou desenvolvimento local:</span><br>
                <strong>http://localhost:8000</strong>
            </div>
        </div>

        <!-- Exemplo cURL -->
        <div class="p-6 bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700">
            <h4 class="text-md font-semibold text-gray-900 dark:text-white mb-2">1. Consulta Individual de Processo (cURL)</h4>
            <div class="bg-gray-900 text-gray-100 p-4 rounded-lg font-mono text-xs overflow-x-auto">
curl -X POST http://localhost:8000/api/v1/processos/consultar \
  -H "Authorization: Bearer rj_live_SEU_TOKEN_AQUI" \
  -H "Content-Type: application/json" \
  -d '{"numero_processo": "0001234-56.2023.8.26.0100"}'
            </div>
        </div>

        <!-- Exemplo Python -->
        <div class="p-6 bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700">
            <h4 class="text-md font-semibold text-gray-900 dark:text-white mb-2">2. Integração em Python (requests)</h4>
            <div class="bg-gray-900 text-gray-100 p-4 rounded-lg font-mono text-xs overflow-x-auto">
import requests

url = "http://localhost:8000/api/v1/processos/consultar"
headers = {
    "Authorization": "Bearer rj_live_SEU_TOKEN_AQUI",
    "Content-Type": "application/json"
}
payload = {
    "numero_processo": "0001234-56.2023.8.26.0100"
}

response = requests.post(url, json=payload, headers=headers)
print("Status:", response.status_code)
print("Dados:", response.json())
            </div>
        </div>

        <!-- Exemplo Node.js -->
        <div class="p-6 bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700">
            <h4 class="text-md font-semibold text-gray-900 dark:text-white mb-2">3. Integração em Node.js / TypeScript (fetch)</h4>
            <div class="bg-gray-900 text-gray-100 p-4 rounded-lg font-mono text-xs overflow-x-auto">
const response = await fetch('http://localhost:8000/api/v1/processos/consultar', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer rj_live_SEU_TOKEN_AQUI',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    numero_processo: '0001234-56.2023.8.26.0100'
  })
});

const data = await response.json();
console.log(data);
            </div>
        </div>

        <!-- Tabela de Status Codes -->
        <div class="p-6 bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700">
            <h4 class="text-md font-semibold text-gray-900 dark:text-white mb-4">Principais Respostas HTTP</h4>
            <div class="overflow-x-auto">
                <table class="w-full text-left text-sm text-gray-600 dark:text-gray-300">
                    <thead class="bg-gray-50 dark:bg-gray-700 text-xs uppercase text-gray-700 dark:text-gray-200">
                        <tr>
                            <th class="p-3">Código</th>
                            <th class="p-3">Significado</th>
                            <th class="p-3">Cobrável na Franquia?</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
                        <tr>
                            <td class="p-3 font-bold text-green-600">200 OK</td>
                            <td class="p-3">Processo localizado e retornado com sucesso.</td>
                            <td class="p-3">Sim</td>
                        </tr>
                        <tr>
                            <td class="p-3 font-bold text-amber-600">400 Bad Request</td>
                            <td class="p-3">Número de processo com formato inválido (deve seguir CNJ NNNNNNN-DD.AAAA.J.TR.OOOO).</td>
                            <td class="p-3">Não</td>
                        </tr>
                        <tr>
                            <td class="p-3 font-bold text-red-600">401 Unauthorized</td>
                            <td class="p-3">Token ausente, inválido ou revogado.</td>
                            <td class="p-3">Não</td>
                        </tr>
                        <tr>
                            <td class="p-3 font-bold text-amber-600">402 Payment Required</td>
                            <td class="p-3">Franquia mensal de consultas esgotada. Faça upgrade do plano.</td>
                            <td class="p-3">Não</td>
                        </tr>
                        <tr>
                            <td class="p-3 font-bold text-amber-600">429 Too Many Requests</td>
                            <td class="p-3">Limite de requisições por minuto excedido (Rate limit ativo).</td>
                            <td class="p-3">Não</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</x-filament-panels::page>
