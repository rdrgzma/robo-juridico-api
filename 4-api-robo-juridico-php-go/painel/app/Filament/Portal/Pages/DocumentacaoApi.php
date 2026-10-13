<?php

namespace App\Filament\Portal\Pages;

use Filament\Actions\Action;
use Filament\Pages\Page;

class DocumentacaoApi extends Page
{
    protected static ?string $navigationIcon = 'heroicon-o-book-open';
    protected static ?string $navigationLabel = 'Guia da API & Docs';
    protected static ?string $title = 'Guia de Integração da API Jurídica';
    protected static ?int $navigationSort = 3;

    protected static string $view = 'filament.portal.pages.documentacao-api';

    protected function getHeaderActions(): array
    {
        return [
            Action::make('abrir_swagger')
                ->label('Abrir Swagger Interativo (Docs)')
                ->icon('heroicon-o-arrow-top-right-on-square')
                ->color('success')
                ->url('http://localhost:8000/docs', shouldOpenInNewTab: true),
        ];
    }
}
