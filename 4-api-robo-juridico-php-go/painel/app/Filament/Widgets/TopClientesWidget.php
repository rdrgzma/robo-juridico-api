<?php

namespace App\Filament\Widgets;

use App\Models\Cliente;
use Carbon\Carbon;
use Filament\Tables;
use Filament\Tables\Table;
use Filament\Widgets\TableWidget as BaseWidget;

class TopClientesWidget extends BaseWidget
{
    protected static ?string $heading = 'Top 5 Clientes em Consumo no Mês';
    protected static ?int $sort = 3;
    protected int | string | array $columnSpan = 'full';

    public function table(Table $table): Table
    {
        $inicioMes = Carbon::now('America/Sao_Paulo')->startOfMonth()->utc()->toIso8601String();

        return $table
            ->query(
                Cliente::query()
                    ->withCount(['usos as consumo_mes' => function ($query) use ($inicioMes) {
                        $query->where('momento', '>=', $inicioMes)
                            ->where('cobravel', 1)
                            ->where(function ($q) {
                                $q->whereNull('status')->orWhere('status', '<', 400);
                            });
                    }])
                    ->orderByDesc('consumo_mes')
                    ->limit(5)
            )
            ->columns([
                Tables\Columns\TextColumn::make('nome')
                    ->label('Cliente')
                    ->weight('bold'),
                Tables\Columns\TextColumn::make('empresa')
                    ->label('Empresa')
                    ->placeholder('—'),
                Tables\Columns\TextColumn::make('plano')
                    ->badge()
                    ->color(fn (string $state): string => match (strtolower($state)) {
                        'enterprise' => 'warning',
                        'pro' => 'success',
                        default => 'gray',
                    }),
                Tables\Columns\TextColumn::make('limite_mensal')
                    ->label('Cota Contratada')
                    ->numeric(),
                Tables\Columns\TextColumn::make('consumo_mes')
                    ->label('Consumo no Mês')
                    ->numeric(),
                Tables\Columns\TextColumn::make('porcentagem')
                    ->label('% da Cota')
                    ->state(function (Cliente $record): string {
                        if ($record->limite_mensal <= 0) {
                            return '0%';
                        }
                        $consumo = $record->consumo_mes ?? 0;
                        $pct = min(100.0, round(($consumo / $record->limite_mensal) * 100, 1));
                        return "{$pct}%";
                    })
                    ->badge()
                    ->color(function (string $state): string {
                        $val = (float) rtrim($state, '%');
                        return match (true) {
                            $val >= 90 => 'danger',
                            $val >= 70 => 'warning',
                            default => 'success',
                        };
                    }),
            ])
            ->paginated(false);
    }
}
