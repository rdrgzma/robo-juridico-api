<?php

namespace App\Filament\Portal\Pages;

use App\Models\Cliente;
use Filament\Forms;
use Filament\Forms\Concerns\InteractsWithForms;
use Filament\Forms\Contracts\HasForms;
use Filament\Forms\Form;
use Filament\Notifications\Notification;
use Filament\Pages\Page;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Hash;

class MeuPlano extends Page implements HasForms
{
    use InteractsWithForms;

    protected static ?string $navigationIcon = 'heroicon-o-identification';
    protected static ?string $navigationLabel = 'Meu Plano & Acesso';
    protected static ?string $title = 'Informações do Meu Plano e Senha';
    protected static ?int $navigationSort = 4;

    protected static string $view = 'filament.portal.pages.meu-plano';

    public ?array $data = [];

    public function mount(): void
    {
        $this->form->fill();
    }

    public function form(Form $form): Form
    {
        return $form
            ->schema([
                Forms\Components\Section::make('Alterar Minha Senha de Acesso')
                    ->description('Atualize com segurança a senha utilizada para entrar neste portal.')
                    ->schema([
                        Forms\Components\TextInput::make('senha_atual')
                            ->label('Senha Atual')
                            ->password()
                            ->required(),
                        Forms\Components\TextInput::make('nova_senha')
                            ->label('Nova Senha')
                            ->password()
                            ->required()
                            ->minLength(6),
                        Forms\Components\TextInput::make('nova_senha_confirmacao')
                            ->label('Confirmar Nova Senha')
                            ->password()
                            ->required()
                            ->same('nova_senha'),
                    ])
                    ->columns(3),
            ])
            ->statePath('data');
    }

    public function submit(): void
    {
        $formData = $this->form->getState();
        /** @var Cliente|null $cliente */
        $cliente = Auth::guard('cliente')->user();

        if (!$cliente || !Hash::check($formData['senha_atual'], $cliente->senha_hash)) {
            Notification::make()
                ->title('Senha atual incorreta!')
                ->danger()
                ->send();
            return;
        }

        $cliente->update([
            'senha_hash' => Hash::make($formData['nova_senha']),
            'atualizado_em' => now()->utc()->toIso8601String(),
        ]);

        $this->form->fill();

        Notification::make()
            ->title('Senha atualizada com sucesso!')
            ->success()
            ->send();
    }

    public function getClienteProperty(): ?Cliente
    {
        return Auth::guard('cliente')->user();
    }
}
