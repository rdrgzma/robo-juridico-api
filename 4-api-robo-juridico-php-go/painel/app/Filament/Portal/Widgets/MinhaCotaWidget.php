<?php

namespace App\Filament\Portal\Widgets;

use App\Models\Cliente;
use Filament\Widgets\StatsOverviewWidget as BaseWidget;
use Filament\Widgets\StatsOverviewWidget\Stat;
use Illuminate\Support\Facades\Auth;

class MinhaCotaWidget extends BaseWidget
{
    protected static ?int $sort = 1;

    protected function getStats(): array
    {
        /** @var Cliente|null $cliente */
        $cliente = Auth::guard('cliente')->user();

        if (!$cliente) {
            return [];
        }

        $consumo = $cliente->consumoMesAtual();
        $limite = $cliente->limite_mensal;
        $saldo = $cliente->saldoRestante();
        $pct = $cliente->porcentagemConsumo();

        $corConsumo = match (true) {
            $pct >= 100.0 => 'danger',
            $pct >= 80.0 => 'warning',
            default => 'success',
        };

        return [
            Stat::make('Consumo no Mês', number_format($consumo, 0, ',', '.') . ' / ' . number_format($limite, 0, ',', '.'))
                ->description("{$pct}% da franquia mensal utilizada")
                ->descriptionIcon('heroicon-m-chart-bar')
                ->color($corConsumo),

            Stat::make('Saldo Restante', number_format($saldo, 0, ',', '.'))
                ->description('Consultas disponíveis até o fim do mês')
                ->descriptionIcon('heroicon-m-check-badge')
                ->color($saldo > 0 ? 'success' : 'danger'),

            Stat::make('Plano Atual', strtoupper($cliente->plano))
                ->description("Vazão máxima: {$cliente->limite_por_minuto} req/min")
                ->descriptionIcon('heroicon-m-bolt')
                ->color('primary'),
        ];
    }
}
