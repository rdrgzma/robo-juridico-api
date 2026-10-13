<x-filament-panels::page>
    <div class="space-y-6">
        @if ($this->cliente)
            <!-- Informações do Contrato / Plano -->
            <div class="p-6 bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700">
                <h3 class="text-lg font-bold text-gray-900 dark:text-white mb-4">Dados Cadastrais & Contratuais</h3>
                
                <div class="grid grid-cols-1 md:grid-cols-3 gap-6 text-sm">
                    <div>
                        <span class="text-gray-500 dark:text-gray-400 block">Razão Social / Nome</span>
                        <strong class="text-gray-900 dark:text-white">{{ $this->cliente->empresa ?: $this->cliente->nome }}</strong>
                    </div>

                    <div>
                        <span class="text-gray-500 dark:text-gray-400 block">E-mail de Login</span>
                        <strong class="text-gray-900 dark:text-white">{{ $this->cliente->email }}</strong>
                    </div>

                    <div>
                        <span class="text-gray-500 dark:text-gray-400 block">Plano Contratado</span>
                        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200 uppercase">
                            {{ $this->cliente->plano }}
                        </span>
                    </div>

                    <div>
                        <span class="text-gray-500 dark:text-gray-400 block">Franquia Mensal</span>
                        <strong class="text-gray-900 dark:text-white">{{ number_format($this->cliente->limite_mensal, 0, ',', '.') }} consultas</strong>
                    </div>

                    <div>
                        <span class="text-gray-500 dark:text-gray-400 block">Taxa Máxima por Minuto</span>
                        <strong class="text-gray-900 dark:text-white">{{ $this->cliente->limite_por_minuto }} req/min</strong>
                    </div>

                    <div>
                        <span class="text-gray-500 dark:text-gray-400 block">Situação da Conta</span>
                        <strong class="{{ $this->cliente->ativo ? 'text-green-600' : 'text-red-600' }}">
                            {{ $this->cliente->ativo ? 'Ativa e Regular' : 'Suspensa' }}
                        </strong>
                    </div>
                </div>
            </div>
        @endif

        <!-- Formulário de Alteração de Senha -->
        <form wire:submit="submit">
            {{ $this->form }}

            <div class="mt-6 flex justify-end">
                <x-filament::button type="submit" color="primary">
                    Atualizar Minha Senha
                </x-filament::button>
            </div>
        </form>
    </div>
</x-filament-panels::page>
