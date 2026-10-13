<?php

namespace App\Filament\Widgets;

use App\Models\Uso;
use Carbon\Carbon;
use Filament\Widgets\ChartWidget;

class ConsultasChartWidget extends ChartWidget
{
    protected static ?string $heading = 'Volume Diário de Consultas (Últimos 14 Dias)';
    protected static ?int $sort = 2;
    protected int | string | array $columnSpan = 'full';

    protected function getData(): array
    {
        $dias = 14;
        $labels = [];
        $sucessos = [];
        $erros = [];

        $hoje = Carbon::now('America/Sao_Paulo');

        for ($i = $dias - 1; $i >= 0; $i--) {
            $data = $hoje->copy()->subDays($i);
            $labels[] = $data->format('d/m');

            $inicioDia = $data->copy()->startOfDay()->utc()->toIso8601String();
            $fimDia = $data->copy()->endOfDay()->utc()->toIso8601String();

            $sucessosCount = Uso::where('momento', '>=', $inicioDia)
                ->where('momento', '<=', $fimDia)
                ->where('status', '>=', 200)
                ->where('status', '<', 400)
                ->count();

            $errosCount = Uso::where('momento', '>=', $inicioDia)
                ->where('momento', '<=', $fimDia)
                ->where('status', '>=', 400)
                ->count();

            $sucessos[] = $sucessosCount;
            $erros[] = $errosCount;
        }

        return [
            'datasets' => [
                [
                    'label' => 'Consultas com Sucesso (2xx/3xx)',
                    'data' => $sucessos,
                    'borderColor' => '#10b981',
                    'backgroundColor' => 'rgba(16, 185, 129, 0.1)',
                    'fill' => true,
                ],
                [
                    'label' => 'Falhas / Bloqueios (4xx/5xx)',
                    'data' => $erros,
                    'borderColor' => '#ef4444',
                    'backgroundColor' => 'rgba(239, 68, 68, 0.1)',
                    'fill' => true,
                ],
            ],
            'labels' => $labels,
        ];
    }

    protected function getType(): string
    {
        return 'line';
    }
}
