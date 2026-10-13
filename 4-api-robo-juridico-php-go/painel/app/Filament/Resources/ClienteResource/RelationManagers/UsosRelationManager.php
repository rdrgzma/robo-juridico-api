<?php

namespace App\Filament\Resources\ClienteResource\RelationManagers;

use App\Models\Uso;
use Filament\Resources\RelationManagers\RelationManager;
use Filament\Tables;
use Filament\Tables\Table;

class UsosRelationManager extends RelationManager
{
    protected static string $relationship = 'usos';
    protected static ?string $title = 'Histórico Recente de Consultas';
    protected static ?string $modelLabel = 'Consulta';

    public function table(Table $table): Table
    {
        return $table
            ->defaultSort('momento', 'desc')
            ->columns([
                Tables\Columns\TextColumn::make('momento_brasilia')
                    ->label('Data / Hora (BRT)')
                    ->sortable(),
                Tables\Columns\TextColumn::make('metodo')
                    ->label('Método')
                    ->badge(),
                Tables\Columns\TextColumn::make('rota')
                    ->label('Endpoint')
                    ->searchable(),
                Tables\Columns\BadgeColumn::make('status')
                    ->label('HTTP')
                    ->colors([
                        'success' => fn ($state) => $state >= 200 && $state < 300,
                        'warning' => fn ($state) => $state >= 400 && $state < 500,
                        'danger' => fn ($state) => $state >= 500,
                    ]),
                Tables\Columns\TextColumn::make('ms')
                    ->label('Duração')
                    ->suffix(' ms'),
                Tables\Columns\IconColumn::make('cobravel')
                    ->label('Faturável')
                    ->boolean(),
                Tables\Columns\TextColumn::make('ip')
                    ->label('IP de Origem'),
            ]);
    }
}
