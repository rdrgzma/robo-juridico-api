<?php

namespace App\Filament\Resources\AdminResource\Pages;

use App\Filament\Resources\AdminResource;
use App\Services\AuditoriaService;
use Filament\Resources\Pages\CreateRecord;
use Illuminate\Support\Facades\Hash;

class CreateAdmin extends CreateRecord
{
    protected static string $resource = AdminResource::class;

    protected function mutateFormDataBeforeCreate(array $data): array
    {
        $data['senha_hash'] = Hash::make($data['senha']);
        unset($data['senha']);
        $data['criado_em'] = now()->utc()->toIso8601String();
        return $data;
    }

    protected function afterCreate(): void
    {
        AuditoriaService::registrar('admin.criado', $this->record->email, "Admin criado: {$this->record->nome}");
    }
}
