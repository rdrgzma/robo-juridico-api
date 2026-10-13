<?php

namespace App\Filament\Resources;

use App\Filament\Resources\ChaveResource\Pages;
use App\Models\Chave;
use App\Services\AuditoriaService;
use Filament\Notifications\Notification;
use Filament\Resources\Resource;
use Filament\Tables;
use Filament\Tables\Table;

class ChaveResource extends Resource
{
    protected static ?string $model = Chave::class;
    protected static ?string $navigationIcon = 'heroicon-o-key';
    protected static ?string $navigationLabel = 'Chaves de API';
    protected static ?string $modelLabel = 'Chave';
    protected static ?string $pluralModelLabel = 'Chaves de API';
    protected static ?int $navigationSort = 2;

    public static function table(Table $table): Table
    {
        return $table
            ->defaultSort('criada_em', 'desc')
            ->columns([
                Tables\Columns\TextColumn::make('cliente.nome')
                    ->label('Cliente')
                    ->searchable()
                    ->sortable(),
                Tables\Columns\TextColumn::make('nome')
                    ->label('Nome da Chave')
                    ->searchable(),
                Tables\Columns\TextColumn::make('prefixo')
                    ->label('Prefixo')
                    ->copyable()
                    ->searchable(),
                Tables\Columns\TextColumn::make('criada_em')
                    ->label('Criada Em')
                    ->sortable(),
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
            ->filters([
                Tables\Filters\SelectFilter::make('cliente')
                    ->relationship('cliente', 'nome'),
                Tables\Filters\TernaryFilter::make('ativa')
                    ->label('Status da Chave')
                    ->nullable()
                    ->queries(
                        true: fn ($query) => $query->whereNull('revogada_em'),
                        false: fn ($query) => $query->whereNotNull('revogada_em'),
                    ),
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
                        AuditoriaService::registrar('chave.revogada', $record->cliente->nome ?? 'Cliente', "Chave {$record->prefixo} revogada");

                        Notification::make()
                            ->title('Chave revogada com sucesso!')
                            ->warning()
                            ->send();
                    }),
            ]);
    }

    public static function getPages(): array
    {
        return [
            'index' => Pages\ListChaves::route('/'),
        ];
    }
}
