<?php

namespace App\Filament\Resources\AdminResource\Pages;

use App\Filament\Resources\AdminResource;
use App\Services\AuditoriaService;
use Filament\Actions;
use Filament\Resources\Pages\EditRecord;
use Illuminate\Support\Facades\Hash;

class EditAdmin extends EditRecord
{
    protected static string $resource = AdminResource::class;

    protected function getHeaderActions(): array
    {
        return [
            Actions\DeleteAction::make(),
        ];
    }

    protected function mutateFormDataBeforeSave(array $data): array
    {
        if (!empty($data['senha'])) {
            $data['senha_hash'] = Hash::make($data['senha']);
        }
        unset($data['senha']);
        return $data;
    }

    protected function afterSave(): void
    {
        AuditoriaService::registrar('admin.atualizado', $this->record->email, "Admin atualizado: {$this->record->nome}");
    }
}
