<?php

namespace App\Filament\Portal\Pages;

use App\Filament\Pages\CalculadoraJudicial as AdminCalculadoraJudicial;

class CalculadoraJudicialPortal extends AdminCalculadoraJudicial
{
    protected static ?string $navigationIcon = 'heroicon-o-calculator';
    protected static ?string $navigationLabel = 'Cálculos & Liquidação';
    protected static ?string $title = 'Calculadora de Débitos Judiciais';
    protected static ?string $navigationGroup = 'Ferramentas';
    protected static ?int $navigationSort = 2;
    protected static ?string $slug = 'calculadora-judicial';
    protected static string $view = 'filament.pages.calculadora-judicial';
}
