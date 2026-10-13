<?php

namespace App\Filament\Resources;

use App\Filament\Resources\ClienteResource\Pages;
use App\Filament\Resources\ClienteResource\RelationManagers;
use App\Models\Cliente;
use App\Services\AuditoriaService;
use App\Services\ChaveService;
use Filament\Forms;
use Filament\Forms\Form;
use Filament\Notifications\Notification;
use Filament\Resources\Resource;
use Filament\Tables;
use Filament\Tables\Table;

class ClienteResource extends Resource
{
    protected static ?string $model = Cliente::class;
    protected static ?string $navigationIcon = 'heroicon-o-users';
    protected static ?string $navigationLabel = 'Clientes & Contratos';
    protected static ?string $modelLabel = 'Cliente';
    protected static ?string $pluralModelLabel = 'Clientes';
    protected static ?int $navigationSort = 1;

    public static function form(Form $form): Form
    {
        return $form->schema([
            Forms\Components\Section::make('Identificação do Cliente')
                ->columns(2)
                ->schema([
                    Forms\Components\TextInput::make('nome')
                        ->label('Nome do Contato')
                        ->required()
                        ->maxLength(255),
                    Forms\Components\TextInput::make('email')
                        ->label('E-mail')
                        ->email()
                        ->required()
                        ->maxLength(255),
                    Forms\Components\TextInput::make('empresa')
                        ->label('Razão Social / Empresa')
                        ->maxLength(255),
                    Forms\Components\Toggle::make('ativo')
                        ->label('Conta Ativa')
                        ->default(true),
                    Forms\Components\Textarea::make('observacoes')
                        ->label('Observações Internas')
                        ->columnSpanFull(),
                ]),

            Forms\Components\Section::make('Plano e Franquia de Consultas')
                ->columns(3)
                ->schema([
                    Forms\Components\Select::make('plano')
                        ->label('Plano Contratado')
                        ->options([
                            'gratuito' => 'Gratuito (Teste)',
                            'basico' => 'Básico',
                            'pro' => 'Profissional',
                            'enterprise' => 'Enterprise',
                            'personalizado' => 'Personalizado',
                        ])
                        ->default('basico')
                        ->required(),
                    Forms\Components\TextInput::make('limite_mensal')
                        ->label('Cota Mensal (Consultas)')
                        ->numeric()
                        ->default(1000)
                        ->required(),
                    Forms\Components\TextInput::make('limite_por_minuto')
                        ->label('Taxa Máxima (Req/min)')
                        ->numeric()
                        ->default(60)
                        ->required(),
                    Forms\Components\TextInput::make('valor_mensal_centavos')
                        ->label('Valor Mensal (Centavos)')
                        ->numeric()
                        ->default(0)
                        ->helperText('Ex: 29900 para R$ 299,00'),
                ]),

            Forms\Components\Section::make('Acesso ao Portal do Cliente')
                ->columns(2)
                ->schema([
                    Forms\Components\TextInput::make('senha')
                        ->password()
                        ->label('Senha do Portal')
                        ->helperText('Defina uma senha para liberar o acesso do cliente ao Portal (/portal).')
                        ->dehydrated(fn ($state) => filled($state))
                        ->maxLength(255),
                    Forms\Components\Placeholder::make('ultimo_login_info')
                        ->label('Último Acesso ao Portal')
                        ->content(fn (?Cliente $record): string => $record?->ultimo_login ?: 'Nunca acessou'),
                ]),
        ]);
    }

    public static function table(Table $table): Table
    {
        return $table
            ->columns([
                Tables\Columns\TextColumn::make('nome')
                    ->label('Cliente')
                    ->searchable()
                    ->sortable(),
                Tables\Columns\TextColumn::make('empresa')
                    ->label('Empresa')
                    ->searchable(),
                Tables\Columns\BadgeColumn::make('plano')
                    ->label('Plano')
                    ->colors([
                        'gray' => 'gratuito',
                        'info' => 'basico',
                        'success' => 'pro',
                        'warning' => 'enterprise',
                    ]),
                Tables\Columns\TextColumn::make('limite_mensal')
                    ->label('Cota Mensal')
                    ->numeric()
                    ->sortable(),
                Tables\Columns\TextColumn::make('consumo_mes')
                    ->label('Consumo no Mês')
                    ->getStateUsing(fn (Cliente $record) => $record->consumoMesAtual() . ' (' . $record->porcentagemConsumo() . '%)'),
                Tables\Columns\IconColumn::make('ativo')
                    ->label('Ativo')
                    ->boolean(),
            ])
            ->actions([
                Tables\Actions\Action::make('gerar_chave')
                    ->label('Nova Chave')
                    ->icon('heroicon-o-key')
                    ->color('success')
                    ->requiresConfirmation()
                    ->form([
                        Forms\Components\TextInput::make('nome_chave')
                            ->label('Identificador da Chave')
                            ->default('Produção')
                            ->required(),
                    ])
                    ->action(function (Cliente $record, array $data) {
                        $service = app(ChaveService::class);
                        $res = $service->gerarParaCliente($record, $data['nome_chave'] ?? 'Nova Chave');
                        $token = $res['token'];

                        AuditoriaService::registrar('chave.gerada', $record->nome, "Chave gerada: {$data['nome_chave']}");

                        Notification::make()
                            ->title('Chave de API Gerada com Sucesso!')
                            ->body("Copie o token agora, ele não será exibido novamente:\n\n{$token}")
                            ->success()
                            ->persistent()
                            ->send();
                    }),
                Tables\Actions\EditAction::make(),
            ])
            ->bulkActions([
                Tables\Actions\BulkActionGroup::make([
                    Tables\Actions\DeleteBulkAction::make(),
                ]),
            ]);
    }

    public static function getRelations(): array
    {
        return [
            RelationManagers\ChavesRelationManager::class,
            RelationManagers\UsosRelationManager::class,
        ];
    }

    public static function getPages(): array
    {
        return [
            'index' => Pages\ListClientes::route('/'),
            'create' => Pages\CreateCliente::route('/create'),
            'edit' => Pages\EditCliente::route('/{record}/edit'),
        ];
    }
}
