<?php

namespace App\Filament\Portal\Resources\MinhaChaveResource\Pages;

use App\Filament\Portal\Resources\MinhaChaveResource;
use App\Models\Cliente;
use App\Services\AuditoriaService;
use App\Services\ChaveService;
use Filament\Actions;
use Filament\Forms;
use Filament\Notifications\Notification;
use Filament\Resources\Pages\ListRecords;
use Illuminate\Support\Facades\Auth;

class ListMinhasChaves extends ListRecords
{
    protected static string $resource = MinhaChaveResource::class;

    protected function getHeaderActions(): array
    {
        return [
            Actions\Action::make('gerar_chave')
                ->label('Gerar Nova Chave de API')
                ->icon('heroicon-o-plus')
                ->color('primary')
                ->form([
                    Forms\Components\TextInput::make('nome')
                        ->label('Identificador da Chave')
                        ->placeholder('Ex: Servidor de Produção, Integração CRM')
                        ->default('Chave de API')
                        ->required()
                        ->maxLength(255),
                ])
                ->action(function (array $data) {
                    /** @var Cliente|null $cliente */
                    $cliente = Auth::guard('cliente')->user();

                    if (!$cliente) {
                        return;
                    }

                    $service = app(ChaveService::class);
                    $resultado = $service->gerarParaCliente($cliente, $data['nome']);
                    $token = $resultado['token'];

                    AuditoriaService::registrar(
                        'chave.gerada_portal',
                        $cliente->nome,
                        "Nova chave emitida pelo desenvolvedor: {$data['nome']}"
                    );

                    Notification::make()
                        ->title('Chave de API Gerada com Sucesso!')
                        ->body("Copie seu token de acesso agora. Por questões de segurança, ele NÃO será exibido novamente:\n\n{$token}")
                        ->success()
                        ->persistent()
                        ->send();
                }),
        ];
    }
}
