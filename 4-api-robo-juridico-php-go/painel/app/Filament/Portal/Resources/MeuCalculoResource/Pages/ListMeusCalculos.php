<?php

namespace App\Filament\Portal\Resources\MeuCalculoResource\Pages;

use App\Filament\Portal\Resources\MeuCalculoResource;
use Filament\Resources\Pages\ListRecords;

class ListMeusCalculos extends ListRecords
{
    protected static string $resource = MeuCalculoResource::class;

    protected function getHeaderActions(): array
    {
        return [];
    }
}
