<?php

namespace App\Filament\Portal\Resources;

use App\Filament\Portal\Resources\MeuCalculoResource\Pages;
use App\Models\CalculoSalvo;
use App\Services\CalculoJudicialService;
use Filament\Notifications\Notification;
use Filament\Resources\Resource;
use Filament\Tables;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Support\Facades\Auth;

class MeuCalculoResource extends Resource
{
    protected static ?string $model = CalculoSalvo::class;
    protected static ?string $navigationIcon = 'heroicon-o-document-duplicate';
    protected static ?string $navigationLabel = 'Meus Laudos Salvos';
    protected static ?string $modelLabel = 'Laudo Judicial';
    protected static ?string $pluralModelLabel = 'Meus Laudos e Liquidações';
    protected static ?string $navigationGroup = 'Ferramentas';
    protected static ?int $navigationSort = 3;

    public static function getEloquentQuery(): Builder
    {
        $cliente = Auth::guard('cliente')->user();
        $clienteId = $cliente?->id ?? 0;
        $email = $cliente?->email ?? '';

        return parent::getEloquentQuery()
            ->where(function (Builder $q) use ($clienteId, $email) {
                $q->where('cliente_id', $clienteId);
                if (!empty($email)) {
                    $q->orWhere('lead_email', $email);
                }
            })
            ->latest('id');
    }

    public static function table(Table $table): Table
    {
        return $table
            ->columns([
                Tables\Columns\TextColumn::make('created_at')
                    ->label('Emitido em')
                    ->dateTime('d/m/Y H:i')
                    ->sortable(),
                Tables\Columns\TextColumn::make('numero_processo')
                    ->label('Processo CNJ')
                    ->fontFamily('mono')
                    ->placeholder('Extrajudicial')
                    ->searchable()
                    ->icon('heroicon-m-arrow-top-right-on-square')
                    ->iconPosition('after')
                    ->color('primary')
                    ->url(function (CalculoSalvo $record): ?string {
                        if (empty($record->numero_processo)) return null;
                        $num = preg_replace('/\D+/', '', (string) $record->numero_processo);
                        return "https://comunica.pje.jus.br/consulta?numeroProcesso={$num}";
                    })
                    ->openUrlInNewTab(),
                Tables\Columns\TextColumn::make('data_atualizacao')
                    ->label('Data-Base')
                    ->date('d/m/Y')
                    ->sortable(),
                Tables\Columns\TextColumn::make('indexador_codigo')
                    ->label('Indexador')
                    ->badge()
                    ->color('primary'),
                Tables\Columns\TextColumn::make('total_geral')
                    ->label('Total Atualizado')
                    ->money('BRL')
                    ->weight('bold')
                    ->sortable(),
            ])
            ->actions([
                Tables\Actions\Action::make('recalcular_hoje')
                    ->label('Atualizar para Hoje (1-Click)')
                    ->icon('heroicon-o-arrow-path')
                    ->color('success')
                    ->requiresConfirmation()
                    ->modalHeading('Atualizar Memória de Cálculo para Hoje')
                    ->modalDescription('Deseja recalcular todas as parcelas deste processo para a data de hoje usando os índices mais recentes e gerar um novo Laudo Certificado?')
                    ->action(function (CalculoSalvo $record, CalculoJudicialService $service) {
                        $novo = $service->atualizarCalculoParaHoje($record);
                        if ($novo) {
                            Notification::make()
                                ->title('Cálculo Atualizado para Hoje!')
                                ->body('Novo total liquidado: R$ ' . number_format((float)$novo->total_geral, 2, ',', '.'))
                                ->success()
                                ->send();
                        } else {
                            Notification::make()
                                ->title('Falha ao Recalcular')
                                ->body('Não foi possível atualizar o cálculo automaticamente.')
                                ->danger()
                                ->send();
                        }
                    }),

                Tables\Actions\Action::make('exportar_xlsx')
                    ->label('Excel (.xlsx)')
                    ->icon('heroicon-o-table-cells')
                    ->color('info')
                    ->url(fn (CalculoSalvo $record): string => route('laudo.exportar', ['uuid' => $record->uuid, 'formato' => 'xlsx'])),

                Tables\Actions\Action::make('imprimir')
                    ->label('Ver Laudo PDF')
                    ->icon('heroicon-o-printer')
                    ->color('gray')
                    ->url(fn (CalculoSalvo $record): string => route('laudo.imprimir', $record->uuid))
                    ->openUrlInNewTab(),
            ]);
    }

    public static function getPages(): array
    {
        return [
            'index' => Pages\ListMeusCalculos::route('/'),
        ];
    }
}
