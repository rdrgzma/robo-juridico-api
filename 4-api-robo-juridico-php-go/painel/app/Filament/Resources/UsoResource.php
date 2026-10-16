<?php

namespace App\Filament\Resources;

use App\Filament\Resources\UsoResource\Pages;
use App\Models\Uso;
use Filament\Resources\Resource;
use Filament\Tables;
use Filament\Tables\Table;

class UsoResource extends Resource
{
    protected static ?string $model = Uso::class;
    protected static ?string $navigationIcon = 'heroicon-o-chart-bar';
    protected static ?string $navigationLabel = 'Auditoria de Consultas';
    protected static ?string $modelLabel = 'Consulta';
    protected static ?string $pluralModelLabel = 'Auditoria de Consultas';
    protected static ?int $navigationSort = 3;

    public static function table(Table $table): Table
    {
        return $table
            ->defaultSort('momento', 'desc')
            ->columns([
                Tables\Columns\TextColumn::make('momento_brasilia')
                    ->label('Data / Hora (BRT)')
                    ->sortable(),
                Tables\Columns\TextColumn::make('cliente.nome')
                    ->label('Cliente')
                    ->searchable()
                    ->sortable()
                    ->placeholder('Chave Interna'),
                Tables\Columns\TextColumn::make('chave.prefixo')
                    ->label('Prefixo Chave')
                    ->placeholder('interna'),
                Tables\Columns\TextColumn::make('metodo')
                    ->label('Método')
                    ->badge(),
                Tables\Columns\TextColumn::make('rota')
                    ->label('Rota')
                    ->getStateUsing(fn ($record) => !empty($record->rota) ? $record->rota : (!empty($record->caminho) ? $record->caminho : '-'))
                    ->searchable()
                    ->placeholder('-'),
                Tables\Columns\TextColumn::make('status')
                    ->label('Status')
                    ->badge()
                    ->colors([
                        'success' => fn ($state) => $state >= 200 && $state < 300,
                        'warning' => fn ($state) => $state >= 400 && $state < 500,
                        'danger' => fn ($state) => $state >= 500,
                    ])
                    ->placeholder('-'),
                Tables\Columns\TextColumn::make('ms')
                    ->label('Latência')
                    ->formatStateUsing(fn ($state) => ($state !== null && $state !== '') ? "{$state} ms" : '-')
                    ->sortable()
                    ->placeholder('-'),
                Tables\Columns\IconColumn::make('cobravel')
                    ->label('Faturável')
                    ->boolean(),
                Tables\Columns\TextColumn::make('ip')
                    ->label('IP')
                    ->searchable(),
            ])
            ->filters([
                Tables\Filters\SelectFilter::make('cliente')
                    ->relationship('cliente', 'nome'),
                Tables\Filters\Filter::make('somente_erros')
                    ->label('Apenas Erros (>= 400)')
                    ->query(fn ($query) => $query->where('status', '>=', 400)),
                Tables\Filters\Filter::make('somente_faturaveis')
                    ->label('Apenas Consultas Faturáveis')
                    ->query(fn ($query) => $query->where('cobravel', 1)),
            ]);
    }

    public static function getPages(): array
    {
        return [
            'index' => Pages\ListUsos::route('/'),
        ];
    }
}
