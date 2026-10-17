<?php

namespace App\Filament\Resources;

use App\Filament\Resources\CalculoSalvoResource\Pages;
use App\Models\CalculoSalvo;
use App\Models\Cliente;
use App\Services\AuditoriaService;
use App\Services\ChaveService;
use Filament\Forms;
use Filament\Forms\Form;
use Filament\Notifications\Notification;
use Filament\Resources\Resource;
use Filament\Tables;
use Filament\Tables\Table;

class CalculoSalvoResource extends Resource
{
    protected static ?string $model = CalculoSalvo::class;
    protected static ?string $navigationIcon = 'heroicon-o-document-check';
    protected static ?string $navigationLabel = 'Laudos & Leads (CRM)';
    protected static ?string $modelLabel = 'Laudo / Lead';
    protected static ?string $pluralModelLabel = 'Laudos & Leads Capturados';
    protected static ?string $navigationGroup = 'Ferramentas Jurídicas';
    protected static ?int $navigationSort = 2;

    public static function form(Form $form): Form
    {
        return $form->schema([
            Forms\Components\Section::make('Dados do Lead e Qualificação Comercial')
                ->columns(2)
                ->schema([
                    Forms\Components\TextInput::make('lead_nome')
                        ->label('Nome / Advogado / Perito')
                        ->maxLength(150),
                    Forms\Components\TextInput::make('lead_email')
                        ->label('E-mail')
                        ->email()
                        ->maxLength(150),
                    Forms\Components\TextInput::make('lead_whatsapp')
                        ->label('WhatsApp')
                        ->maxLength(30),
                    Forms\Components\TextInput::make('lead_oab')
                        ->label('Inscrição OAB / Órgão')
                        ->maxLength(30),
                    Forms\Components\Select::make('status_lead')
                        ->label('Status Comercial (Funil CRM)')
                        ->options([
                            'novo' => 'Novo Lead',
                            'contatado' => 'Contatado via WhatsApp',
                            'em_negociacao' => 'Em Negociação',
                            'convertido' => 'Convertido em Cliente',
                        ])
                        ->default('novo')
                        ->required(),
                    Forms\Components\TextInput::make('numero_processo')
                        ->label('Processo CNJ')
                        ->disabled(),
                    Forms\Components\Textarea::make('observacoes_crm')
                        ->label('Anotações do Comercial / Follow-up')
                        ->columnSpanFull(),
                ]),
        ]);
    }

    public static function table(Table $table): Table
    {
        return $table
            ->defaultSort('created_at', 'desc')
            ->columns([
                Tables\Columns\TextColumn::make('created_at')
                    ->label('Emitido em')
                    ->dateTime('d/m/Y H:i')
                    ->sortable(),
                Tables\Columns\TextColumn::make('lead_nome')
                    ->label('Profissional / Lead')
                    ->placeholder('Anônimo / Painel')
                    ->searchable()
                    ->weight('bold'),
                Tables\Columns\TextColumn::make('lead_oab')
                    ->label('OAB')
                    ->searchable()
                    ->placeholder('-'),
                Tables\Columns\TextColumn::make('lead_whatsapp')
                    ->label('WhatsApp')
                    ->searchable()
                    ->placeholder('-'),
                Tables\Columns\TextColumn::make('numero_processo')
                    ->label('Processo CNJ')
                    ->fontFamily('mono')
                    ->searchable()
                    ->placeholder('Extrajudicial')
                    ->icon('heroicon-m-arrow-top-right-on-square')
                    ->iconPosition('after')
                    ->color('primary')
                    ->url(function (CalculoSalvo $record): ?string {
                        if (empty($record->numero_processo)) return null;
                        $num = preg_replace('/\D+/', '', (string) $record->numero_processo);
                        return "https://comunica.pje.jus.br/consulta?numeroProcesso={$num}";
                    })
                    ->openUrlInNewTab(),
                Tables\Columns\TextColumn::make('indexador_codigo')
                    ->label('Indexador')
                    ->badge()
                    ->color('gray'),
                Tables\Columns\TextColumn::make('total_geral')
                    ->label('Valor Liquidado')
                    ->money('BRL')
                    ->sortable(),
                Tables\Columns\BadgeColumn::make('status_lead')
                    ->label('Status CRM')
                    ->colors([
                        'info' => 'novo',
                        'warning' => 'contatado',
                        'primary' => 'em_negociacao',
                        'success' => 'convertido',
                    ]),
            ])
            ->filters([
                Tables\Filters\SelectFilter::make('status_lead')
                    ->label('Status do Lead')
                    ->options([
                        'novo' => 'Novo Lead',
                        'contatado' => 'Contatado',
                        'em_negociacao' => 'Em Negociação',
                        'convertido' => 'Convertido',
                    ]),
            ])
            ->actions([
                Tables\Actions\Action::make('ver_laudo')
                    ->label('Laudo PDF')
                    ->icon('heroicon-o-printer')
                    ->color('gray')
                    ->url(fn (CalculoSalvo $record): string => route('laudo.imprimir', $record->uuid))
                    ->openUrlInNewTab(),

                Tables\Actions\Action::make('whatsapp')
                    ->label('WhatsApp')
                    ->icon('heroicon-o-chat-bubble-left-ellipsis')
                    ->color('success')
                    ->visible(fn (CalculoSalvo $record): bool => !empty($record->lead_whatsapp))
                    ->url(function (CalculoSalvo $record): string {
                        $fone = preg_replace('/\D+/', '', (string) $record->lead_whatsapp);
                        if (strlen($fone) <= 11 && !str_starts_with($fone, '55')) {
                            $fone = '55' . $fone;
                        }
                        $msg = urlencode("Olá {$record->lead_nome}, vimos que você emitiu um Laudo de Liquidação Judicial no Robô Jurídico. Posso liberar uma chave de testes da nossa API para o seu escritório?");
                        return "https://wa.me/{$fone}?text={$msg}";
                    })
                    ->openUrlInNewTab(),

                Tables\Actions\Action::make('converter_cliente')
                    ->label('Converter em Cliente')
                    ->icon('heroicon-o-user-plus')
                    ->color('primary')
                    ->visible(fn (CalculoSalvo $record): bool => $record->status_lead !== 'convertido' && !empty($record->lead_email))
                    ->requiresConfirmation()
                    ->modalHeading('Converter Lead em Assinante da API')
                    ->modalDescription('Cria automaticamente um cadastro de Cliente com plano Básico, vincula este cálculo e gera a primeira Chave de API.')
                    ->action(function (CalculoSalvo $record) {
                        $cliente = Cliente::firstOrCreate(
                            ['email' => $record->lead_email],
                            [
                                'nome' => $record->lead_nome ?: 'Novo Cliente Lead',
                                'empresa' => $record->lead_oab ?: 'Escritório de Advocacia',
                                'plano' => 'basico',
                                'limite_mensal' => 1000,
                                'limite_por_minuto' => 60,
                                'valor_mensal_centavos' => 9900,
                                'ativo' => true,
                                'observacoes' => "Convertido via Lead Magnet (Laudo #{$record->uuid})",
                            ]
                        );

                        $record->update([
                            'cliente_id' => $cliente->id,
                            'status_lead' => 'convertido',
                        ]);

                        $chaveService = app(ChaveService::class);
                        $res = $chaveService->gerarParaCliente($cliente, 'Chave Inicial Lead');

                        AuditoriaService::registrar('lead.convertido', $cliente->nome, "Lead convertido em cliente #{$cliente->id}");

                        Notification::make()
                            ->title('Lead Convertido em Cliente!')
                            ->body("Cliente '{$cliente->nome}' criado. Token gerado:\n{$res['token']}")
                            ->success()
                            ->persistent()
                            ->send();
                    }),

                Tables\Actions\EditAction::make()->label('Qualificar'),
            ]);
    }

    public static function getPages(): array
    {
        return [
            'index' => Pages\ListCalculosSalvos::route('/'),
        ];
    }
}
