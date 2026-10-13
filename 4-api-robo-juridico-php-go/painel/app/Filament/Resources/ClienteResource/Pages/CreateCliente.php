<?php

namespace App\Filament\Resources\ClienteResource\Pages;

use App\Filament\Resources\ClienteResource;
use App\Services\AuditoriaService;
use Filament\Resources\Pages\CreateRecord;

class CreateCliente extends CreateRecord
{
    protected static string $resource = ClienteResource::class;

    protected function mutateFormDataBeforeCreate(array $data): array
    {
        $agora = now()->utc()->toIso8601String();
        $data['criado_em'] = $agora;
        $data['atualizado_em'] = $agora;
        return $data;
    }

    protected function afterCreate(): void
    {
        AuditoriaService::registrar('cliente.criado', $this->record->nome, "Plano: {$this->record->plano}");
    }
}
