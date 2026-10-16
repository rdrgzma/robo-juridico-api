<?php

namespace App\Filament\Widgets;

use App\Models\CalculoSalvo;
use App\Models\Cliente;
use App\Models\Uso;
use Carbon\Carbon;
use Filament\Widgets\StatsOverviewWidget as BaseWidget;
use Filament\Widgets\StatsOverviewWidget\Stat;

class StatsOverviewWidget extends BaseWidget
{
    protected static ?int $sort = 1;

    protected function getStats(): array
    {
        $inicioMes = Carbon::now('America/Sao_Paulo')->startOfMonth()->utc()->toIso8601String();

        $totalConsultasMes = Uso::where('momento', '>=', $inicioMes)->count();
        $totalClientesAtivos = Cliente::where('ativo', true)->count();
        $mrrCentavos = (int) Cliente::where('ativo', true)->sum('valor_mensal_centavos');
        $mrrEstimado = $mrrCentavos / 100;
        $bloqueiosCota = Uso::where('status', 402)
            ->where('momento', '>=', $inicioMes)
            ->count();
        $totalLeads = CalculoSalvo::count();

        return [
            Stat::make('Consultas no Mês', number_format($totalConsultasMes, 0, ',', '.'))
                ->description('Total de requisições registradas')
                ->descriptionIcon('heroicon-m-arrow-trending-up')
                ->color('primary'),

            Stat::make('Clientes Ativos', $totalClientesAtivos)
                ->description('Empresas com plano ativo')
                ->descriptionIcon('heroicon-m-user-group')
                ->color('success'),

            Stat::make('MRR Estimado', 'R$ ' . number_format($mrrEstimado, 2, ',', '.'))
                ->description('Receita recorrente mensal contratada')
                ->descriptionIcon('heroicon-m-banknotes')
                ->color('warning'),

            Stat::make('Laudos & Leads (CRM)', $totalLeads)
                ->description($bloqueiosCota > 0 ? "{$bloqueiosCota} bloqueios de cota (402)" : 'Capturados na Calculadora Pública')
                ->descriptionIcon('heroicon-m-document-check')
                ->color('info'),
        ];
    }
}
