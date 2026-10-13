<?php

namespace App\Filament\Resources\ClienteResource\Pages;

use App\Filament\Resources\ClienteResource;
use App\Services\AuditoriaService;
use Filament\Actions;
use Filament\Resources\Pages\EditRecord;

class EditCliente extends EditRecord
{
    protected static string $resource = ClienteResource::class;

    protected function getHeaderActions(): array
    {
        return [
            Actions\DeleteAction::make(),
        ];
    }

    protected function mutateFormDataBeforeSave(array $data): array
    {
        $data['atualizado_em'] = now()->utc()->toIso8601String();
        return $data;
    }

    protected function afterSave(): void
    {
        AuditoriaService::registrar('cliente.editado', $this->record->nome, "Cota: {$this->record->limite_mensal}");
    }
}
