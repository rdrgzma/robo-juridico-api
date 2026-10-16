<?php

namespace App\Filament\Portal\Resources;

use App\Filament\Portal\Resources\MeuUsoResource\Pages;
use App\Models\Uso;
use Filament\Forms\Form;
use Filament\Resources\Resource;
use Filament\Tables;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Support\Facades\Auth;

class MeuUsoResource extends Resource
{
    protected static ?string $model = Uso::class;
    protected static ?string $navigationIcon = 'heroicon-o-chart-bar-square';
    protected static ?string $navigationLabel = 'Histórico de Consultas';
    protected static ?string $modelLabel = 'Consulta';
    protected static ?string $pluralModelLabel = 'Histórico de Consultas';
    protected static ?int $navigationSort = 2;

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
                Tables\Columns\TextColumn::make('momento_brasilia')
                    ->label('Data / Hora (Brasília)')
                    ->sortable(),
                Tables\Columns\TextColumn::make('metodo')
                    ->label('Método')
                    ->badge()
                    ->colors([
                        'primary' => 'GET',
                        'success' => 'POST',
                    ]),
                Tables\Columns\TextColumn::make('rota')
                    ->label('Endpoint')
                    ->getStateUsing(fn ($record) => !empty($record->rota) ? $record->rota : (!empty($record->caminho) ? $record->caminho : '-'))
                    ->searchable()
                    ->fontFamily('mono')
                    ->placeholder('-'),
                Tables\Columns\TextColumn::make('status')
                    ->label('Status')
                    ->badge()
                    ->colors([
                        'success' => fn ($state): bool => (int) $state >= 200 && (int) $state < 300,
                        'warning' => fn ($state): bool => (int) $state === 402,
                        'danger' => fn ($state): bool => (int) $state >= 400 && (int) $state !== 402,
                    ])
                    ->placeholder('-'),
                Tables\Columns\TextColumn::make('ms')
                    ->label('Latência')
                    ->formatStateUsing(fn ($state) => ($state !== null && $state !== '') ? "{$state} ms" : '-')
                    ->sortable()
                    ->placeholder('-'),
                Tables\Columns\IconColumn::make('cobravel')
                    ->label('Cobrável')
                    ->boolean(),
                Tables\Columns\TextColumn::make('ip')
                    ->label('IP de Origem')
                    ->fontFamily('mono')
                    ->toggleable(isToggledHiddenByDefault: true),
            ])
            ->filters([
                Tables\Filters\SelectFilter::make('status')
                    ->label('Status HTTP')
                    ->options([
                        '200' => '200 OK (Sucesso)',
                        '400' => '400 Bad Request',
                        '401' => '401 Não Autorizado',
                        '402' => '402 Cota Excedida',
                        '404' => '404 Não Encontrado',
                        '429' => '429 Rate Limit Excedido',
                        '500' => '500 Erro Interno',
                    ]),
            ])
            ->defaultSort('id', 'desc');
    }

    public static function getPages(): array
    {
        return [
            'index' => Pages\ListMeusUsos::route('/'),
        ];
    }
}
