<?php

namespace App\Filament\Portal\Widgets;

use App\Models\Chave;
use App\Models\Cliente;
use Filament\Tables;
use Filament\Tables\Table;
use Filament\Widgets\TableWidget as BaseWidget;
use Illuminate\Support\Facades\Auth;

class ChavesAtivasWidget extends BaseWidget
{
    protected static ?string $heading = 'Minhas Chaves de API Ativas';
    protected static ?int $sort = 3;
    protected int | string | array $columnSpan = 'full';

    public function table(Table $table): Table
    {
        /** @var Cliente|null $cliente */
        $cliente = Auth::guard('cliente')->user();
        $clienteId = $cliente?->id ?? 0;

        return $table
            ->query(
                Chave::query()
                    ->where('cliente_id', $clienteId)
                    ->whereNull('revogada_em')
                    ->latest('id')
            )
            ->columns([
                Tables\Columns\TextColumn::make('nome')
                    ->label('Identificador')
                    ->weight('bold'),
                Tables\Columns\TextColumn::make('prefixo')
                    ->label('Prefixo do Token')
                    ->fontFamily('mono')
                    ->copyable(),
                Tables\Columns\TextColumn::make('criada_em')
                    ->label('Data de Criação'),
                Tables\Columns\TextColumn::make('ultimo_uso')
                    ->label('Última Utilização')
                    ->placeholder('Ainda não utilizada'),
            ])
            ->paginated(false);
    }
}
