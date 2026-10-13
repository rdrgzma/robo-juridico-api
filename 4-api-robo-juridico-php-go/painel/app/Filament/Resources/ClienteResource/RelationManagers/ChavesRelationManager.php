<?php

namespace App\Filament\Resources\ClienteResource\RelationManagers;

use App\Models\Chave;
use App\Services\AuditoriaService;
use Filament\Forms;
use Filament\Forms\Form;
use Filament\Notifications\Notification;
use Filament\Resources\RelationManagers\RelationManager;
use Filament\Tables;
use Filament\Tables\Table;

class ChavesRelationManager extends RelationManager
{
    protected static string $relationship = 'chaves';
    protected static ?string $title = 'Chaves de API';
    protected static ?string $modelLabel = 'Chave';

    public function form(Form $form): Form
    {
        return $form->schema([
            Forms\Components\TextInput::make('nome')
                ->required()
                ->maxLength(255),
        ]);
    }

    public function table(Table $table): Table
    {
        return $table
            ->columns([
                Tables\Columns\TextColumn::make('nome')
                    ->label('Identificador')
                    ->searchable(),
                Tables\Columns\TextColumn::make('prefixo')
                    ->label('Prefixo do Token')
                    ->copyable(),
                Tables\Columns\TextColumn::make('criada_em')
                    ->label('Gerada em')
                    ->dateTime('d/m/Y H:i'),
                Tables\Columns\TextColumn::make('ultimo_uso')
                    ->label('Último Uso')
                    ->placeholder('Nunca utilizada'),
                Tables\Columns\BadgeColumn::make('status')
                    ->label('Status')
                    ->getStateUsing(fn (Chave $record) => $record->isAtiva() ? 'Ativa' : 'Revogada')
                    ->colors([
                        'success' => 'Ativa',
                        'danger' => 'Revogada',
                    ]),
            ])
            ->actions([
                Tables\Actions\Action::make('revogar')
                    ->label('Revogar')
                    ->icon('heroicon-o-x-circle')
                    ->color('danger')
                    ->requiresConfirmation()
                    ->visible(fn (Chave $record) => $record->isAtiva())
                    ->action(function (Chave $record) {
                        $record->revogar();
                        AuditoriaService::registrar('chave.revogada', $record->cliente->nome, "Chave {$record->prefixo} revogada");

                        Notification::make()
                            ->title('Chave revogada com sucesso!')
                            ->warning()
                            ->send();
                    }),
            ]);
    }
}
