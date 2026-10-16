<?php

namespace App\Filament\Resources\CalculoSalvoResource\Pages;

use App\Filament\Resources\CalculoSalvoResource;
use Filament\Resources\Pages\ListRecords;

class ListCalculosSalvos extends ListRecords
{
    protected static string $resource = CalculoSalvoResource::class;

    protected function getHeaderActions(): array
    {
        return [];
    }
}
