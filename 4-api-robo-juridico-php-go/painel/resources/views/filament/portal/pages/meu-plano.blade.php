<x-filament-panels::page>
    <div class="space-y-6">
        @if ($this->cliente)
            <!-- Informações do Contrato / Plano -->
            <x-filament::section>
                <x-slot name="heading">
                    Dados Cadastrais & Contratuais
                </x-slot>

                <div class="grid grid-cols-1 md:grid-cols-3 gap-6 text-sm">
                    <div>
                        <span class="text-gray-500 dark:text-gray-400 block text-xs uppercase font-medium">Razão Social / Nome</span>
                        <strong class="text-gray-900 dark:text-white font-semibold">{{ $this->cliente->empresa ?: $this->cliente->nome }}</strong>
                    </div>

                    <div>
                        <span class="text-gray-500 dark:text-gray-400 block text-xs uppercase font-medium">E-mail de Login</span>
                        <strong class="text-gray-900 dark:text-white font-semibold">{{ $this->cliente->email }}</strong>
                    </div>

                    <div>
                        <span class="text-gray-500 dark:text-gray-400 block text-xs uppercase font-medium mb-1">Plano Contratado</span>
                        <x-filament::badge color="success">
                            {{ strtoupper($this->cliente->plano) }}
                        </x-filament::badge>
                    </div>

                    <div>
                        <span class="text-gray-500 dark:text-gray-400 block text-xs uppercase font-medium">Franquia Mensal</span>
                        <strong class="text-gray-900 dark:text-white font-semibold">{{ number_format($this->cliente->limite_mensal, 0, ',', '.') }} consultas</strong>
                    </div>

                    <div>
                        <span class="text-gray-500 dark:text-gray-400 block text-xs uppercase font-medium">Taxa Máxima por Minuto</span>
                        <strong class="text-gray-900 dark:text-white font-semibold">{{ $this->cliente->limite_por_minuto }} req/min</strong>
                    </div>

                    <div>
                        <span class="text-gray-500 dark:text-gray-400 block text-xs uppercase font-medium mb-1">Situação da Conta</span>
                        <x-filament::badge :color="$this->cliente->ativo ? 'success' : 'danger'">
                            {{ $this->cliente->ativo ? 'Ativa e Regular' : 'Suspensa' }}
                        </x-filament::badge>
                    </div>
                </div>
            </x-filament::section>
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
