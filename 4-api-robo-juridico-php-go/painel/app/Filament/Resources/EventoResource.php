<?php

namespace App\Filament\Resources;

use App\Filament\Resources\EventoResource\Pages;
use App\Models\Evento;
use Filament\Resources\Resource;
use Filament\Tables;
use Filament\Tables\Table;

class EventoResource extends Resource
{
    protected static ?string $model = Evento::class;
    protected static ?string $navigationIcon = 'heroicon-o-shield-check';
    protected static ?string $navigationLabel = 'Auditoria do Painel';
    protected static ?string $modelLabel = 'Evento';
    protected static ?string $pluralModelLabel = 'Auditoria do Painel';
    protected static ?int $navigationSort = 4;

    public static function table(Table $table): Table
    {
        return $table
            ->defaultSort('momento', 'desc')
            ->columns([
                Tables\Columns\TextColumn::make('momento_brasilia')
                    ->label('Data / Hora (BRT)')
                    ->sortable(),
                Tables\Columns\TextColumn::make('admin_email')
                    ->label('Administrador')
                    ->searchable(),
                Tables\Columns\BadgeColumn::make('acao')
                    ->label('Ação Executada')
                    ->colors([
                        'primary' => fn ($state) => str_contains($state, 'criado'),
                        'warning' => fn ($state) => str_contains($state, 'editado'),
                        'danger' => fn ($state) => str_contains($state, 'revogada'),
                    ]),
                Tables\Columns\TextColumn::make('alvo')
                    ->label('Alvo')
                    ->searchable(),
                Tables\Columns\TextColumn::make('detalhe')
                    ->label('Detalhes')
                    ->wrap(),
                Tables\Columns\TextColumn::make('ip')
                    ->label('IP'),
            ]);
    }

    public static function getPages(): array
    {
        return [
            'index' => Pages\ListEventos::route('/'),
        ];
    }
}
