<?php

namespace App\Filament\Portal\Resources;

use App\Filament\Portal\Resources\MinhaChaveResource\Pages;
use App\Models\Chave;
use App\Services\AuditoriaService;
use Filament\Forms\Form;
use Filament\Notifications\Notification;
use Filament\Resources\Resource;
use Filament\Tables;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Support\Facades\Auth;

class MinhaChaveResource extends Resource
{
    protected static ?string $model = Chave::class;
    protected static ?string $navigationIcon = 'heroicon-o-key';
    protected static ?string $navigationLabel = 'Minhas Chaves de API';
    protected static ?string $modelLabel = 'Chave de API';
    protected static ?string $pluralModelLabel = 'Minhas Chaves de API';
    protected static ?int $navigationSort = 1;

    public static function getEloquentQuery(): Builder
    {
        $clienteId = Auth::guard('cliente')->id() ?? 0;

        return parent::getEloquentQuery()
            ->where('cliente_id', $clienteId)
            ->latest('id');
    }

    public static function table(Table $table): Table
    {
        return $table
            ->columns([
                Tables\Columns\TextColumn::make('nome')
                    ->label('Identificador')
                    ->weight('bold')
                    ->searchable(),
                Tables\Columns\TextColumn::make('prefixo')
                    ->label('Prefixo')
                    ->fontFamily('mono')
                    ->copyable()
                    ->copyMessage('Prefixo copiado com sucesso!'),
                Tables\Columns\TextColumn::make('criada_em')
                    ->label('Emitida em')
                    ->sortable(),
                Tables\Columns\TextColumn::make('ultimo_uso')
                    ->label('Último Uso')
                    ->placeholder('Nunca utilizada'),
                Tables\Columns\TextColumn::make('status')
                    ->label('Status')
                    ->state(fn (Chave $record): string => $record->isAtiva() ? 'Ativa' : 'Revogada')
                    ->badge()
                    ->color(fn (string $state): string => $state === 'Ativa' ? 'success' : 'danger'),
            ])
            ->actions([
                Tables\Actions\Action::make('revogar')
                    ->label('Revogar')
                    ->icon('heroicon-o-x-circle')
                    ->color('danger')
                    ->visible(fn (Chave $record): bool => $record->isAtiva())
                    ->requiresConfirmation()
                    ->modalHeading('Revogar Chave de API')
                    ->modalDescription('Tem certeza que deseja revogar esta chave? Qualquer integração ou sistema utilizando este token deixará de funcionar imediatamente.')
                    ->action(function (Chave $record) {
                        $record->revogar();

                        $cliente = Auth::guard('cliente')->user();
                        AuditoriaService::registrar(
                            'chave.revogada_portal',
                            $cliente?->nome ?: 'Cliente',
                            "Chave revogada pelo desenvolvedor: {$record->nome}"
                        );

                        Notification::make()
                            ->title('Chave Revogada com Sucesso')
                            ->success()
                            ->send();
                    }),
            ]);
    }

    public static function getPages(): array
    {
        return [
            'index' => Pages\ListMinhasChaves::route('/'),
        ];
    }
}
