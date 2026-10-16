<?php

namespace App\Filament\Pages;

use App\Services\CalculoJudicialService;
use Filament\Actions\Action;
use Filament\Forms\Components\DatePicker;
use Filament\Forms\Components\FileUpload;
use Filament\Forms\Components\Grid;
use Filament\Forms\Components\Radio;
use Filament\Forms\Components\Repeater;
use Filament\Forms\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\Textarea;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Forms\Concerns\InteractsWithForms;
use Filament\Forms\Contracts\HasForms;
use Filament\Forms\Form;
use Filament\Notifications\Notification;
use Filament\Pages\Page;
use Illuminate\Support\Facades\Storage;

class CalculadoraJudicial extends Page implements HasForms
{
    use InteractsWithForms;

    protected static ?string $navigationIcon = 'heroicon-o-calculator';
    protected static ?string $navigationLabel = 'Cálculos Judiciais';
    protected static ?string $title = 'Calculadora de Débitos Judiciais';
    protected static ?string $navigationGroup = 'Ferramentas Jurídicas';
    protected static ?int $navigationSort = 1;
    protected static string $view = 'filament.pages.calculadora-judicial';

    public ?array $data = [];
    public ?array $resultado = null;
    public ?array $dadosProcesso = null;

    public function mount(): void
    {
        $this->form->fill([
            'data_atualizacao' => now()->format('Y-m-d'),
            'indexador_codigo' => 'TJSP',
            'aplicar_deflacao' => false,
            'juros_tipo' => 'CONVENCIONAL_MENSAL',
            'juros_percentual' => 1.0,
            'juros_modo' => 'COMPOSTO',
            'juros_prorata' => true,
            'aplicar_art_523_multa' => true,
            'aplicar_art_523_honorarios' => true,
            'multa_moratoria_percentual' => 0.0,
            'honorarios_sucumbenciais_percentual' => 0.0,
            'parcelas' => [
                [
                    'descricao' => 'Parcela 01 - Principal',
                    'tipo' => 'DEBITO',
                    'data_vencimento' => now()->subYears(2)->format('Y-m-d'),
                    'valor_original' => 10000.00,
                ],
            ],
        ]);
    }

    public function form(Form $form): Form
    {
        return $form
            ->schema([
                Section::make('Identificação do Processo')
                    ->description('Informe o número CNJ para auto-identificação do Tribunal e das partes.')
                    ->collapsible()
                    ->schema([
                        Grid::make(3)->schema([
                            TextInput::make('numero_processo')
                                ->label('Número do Processo (CNJ)')
                                ->placeholder('0000000-00.0000.0.00.0000')
                                ->suffixAction(
                                    \Filament\Forms\Components\Actions\Action::make('identificar')
                                        ->icon('heroicon-m-sparkles')
                                        ->tooltip('Auto-Preencher Tribunal e Dados pelo CNJ (1-Click)')
                                        ->action(function ($state, CalculoJudicialService $service, \Filament\Forms\Set $set, \Filament\Forms\Get $get) {
                                            if (empty($state)) return;
                                            $info = $service->autocompletarPorProcesso($state);
                                            if (!empty($info['ok'])) {
                                                $this->dadosProcesso = $info;
                                                $set('numero_processo', $info['numero_formatado']);
                                                if (!empty($info['indexador_sugerido'])) {
                                                    $set('indexador_codigo', $info['indexador_sugerido']);
                                                }
                                                $parcelasAtuais = $get('parcelas') ?? [];
                                                if (count($parcelasAtuais) <= 1) {
                                                    $set('parcelas', [[
                                                        'descricao' => $info['descricao_sugerida'] ?? 'Valor Principal da Condenação',
                                                        'tipo' => 'DEBITO',
                                                        'data_vencimento' => $info['data_sugerida'] ?? now()->subYear()->format('Y-m-d'),
                                                        'valor_original' => $info['valor_sugerido'] ?? ($parcelasAtuais[0]['valor_original'] ?? 10000.00),
                                                    ]]);
                                                }
                                                Notification::make()
                                                    ->title("Processo Identificado ({$info['tribunal']})")
                                                    ->body("Indexador ajustado para {$info['indexador_sugerido']} | Segmento: {$info['segmento']}")
                                                    ->success()
                                                    ->send();
                                            } else {
                                                Notification::make()
                                                    ->title('Número CNJ Inválido')
                                                    ->body($info['mensagem'] ?? 'Verifique a numeração informada.')
                                                    ->warning()
                                                    ->send();
                                            }
                                        })
                                ),
                            Select::make('indexador_codigo')
                                ->label('Indexador de Correção')
                                ->required()
                                ->searchable()
                                ->options([
                                    'Tribunais de Justiça Estaduais' => [
                                        'TJSP' => 'TJSP - Tabela Prática Débitos Judiciais (Lei 14.905)',
                                        'TJRJ' => 'TJRJ - Tabela Tribunal de Justiça do Rio de Janeiro',
                                        'TJMG' => 'TJMG - Tabela Tribunal de Justiça de MG (Não Expurgada)',
                                        'TJMG_EXPURGADA' => 'TJMG - Tabela Tribunal de Justiça de MG (Expurgada)',
                                        'TJRS' => 'TJRS - Tabela Tribunal de Justiça do Rio Grande do Sul',
                                        'TJPR' => 'TJPR - Critério Oficial Média (IGP-DI / INPC)',
                                        'TJPR_PRECAT' => 'TJPR - Tabela de Precatórios (IPCA-E)',
                                        'TJSC' => 'TJSC - Tabela ICGJ de Santa Catarina',
                                        'TJES' => 'TJES - Tabela Tribunal de Justiça do Espírito Santo',
                                        'TJCE' => 'TJCE - Tabela Tribunal de Justiça do Ceará',
                                        'TJDF' => 'TJDF - Tabela Tribunal de Justiça do Distrito Federal',
                                    ],
                                    'Justiça Federal e Especializadas' => [
                                        'JF_GERAL' => 'Justiça Federal - Condenatórias em Geral (Res. 267/2013)',
                                        'JF_PREVIDENCIARIO' => 'Justiça Federal - Benefício Previdenciário (INPC)',
                                        'JF_FAZENDA' => 'Justiça Federal - Fazenda Pública',
                                        'JF_DESAPROPRIACAO' => 'Justiça Federal - Desapropriações (Res. 267/2013)',
                                        'TST_TRABALHISTA' => 'TST - Débitos Trabalhistas (ADC 58 STF)',
                                    ],
                                    'Índices de Preços e Inflação' => [
                                        'IPCA' => 'IPCA (IBGE - Padrão Código Civil / Lei 14.905)',
                                        'INPC' => 'INPC (IBGE - Trabalhista / Previdenciário)',
                                        'IPCA_15' => 'IPCA-15 (IBGE)',
                                        'IGPM' => 'IGP-M (FGV - Reajuste de Locações)',
                                        'IGP_DI' => 'IGP-DI (FGV)',
                                        'INCC' => 'INCC-DI (FGV - Custo da Construção)',
                                    ],
                                    'Taxas Financeiras e Legais' => [
                                        'TAXA_LEGAL' => 'Taxa Legal de Juros (Lei 14.905/2024 - Bacen)',
                                        'SELIC' => 'Taxa SELIC Acumulada Mensal (% a.m.)',
                                        'CDI' => 'CDI Acumulado Mensal (% a.m.)',
                                        'TR' => 'Taxa Referencial (TR)',
                                        'POUPANCA' => 'Poupança Nova (% a.m.)',
                                    ],
                                ])
                                ->default('TJSP'),
                            DatePicker::make('data_atualizacao')
                                ->label('Data da Atualização (Termo Ad Quem)')
                                ->required()
                                ->default(now()->format('Y-m-d')),
                        ]),
                        Grid::make(3)->schema([
                            Toggle::make('aplicar_deflacao')
                                ->label('Aplicar Deflação (Inflação negativa)')
                                ->default(false),
                            Toggle::make('aplicar_prorata_correcao')
                                ->label('Correção Monetária Pro Rata Die')
                                ->default(false),
                            Toggle::make('imputar_art_354_cc')
                                ->label('Imputação Art. 354 CC (Abater Juros Primeiro)')
                                ->default(false),
                        ]),
                    ]),

                Section::make('Parâmetros de Juros Moratórios, Compensatórios e Execução')
                    ->collapsible()
                    ->schema([
                        Grid::make(4)->schema([
                            Select::make('juros_tipo')
                                ->label('Regime de Juros')
                                ->options([
                                    'CONVENCIONAL_MENSAL' => 'Convencional (ex: 1% a.m.)',
                                    'TAXA_LEGAL' => 'Taxa Legal (Lei 14.905/2024 - Bacen)',
                                    'SELIC' => 'Taxa SELIC (Juros + Correção)',
                                    'SEM_JUROS' => 'Sem Incidência de Juros',
                                ])
                                ->default('CONVENCIONAL_MENSAL')
                                ->live(),
                            TextInput::make('juros_percentual')
                                ->label('Taxa Mensal (%)')
                                ->numeric()
                                ->default(1.0)
                                ->visible(fn ($get) => $get('juros_tipo') === 'CONVENCIONAL_MENSAL'),
                            Radio::make('juros_modo')
                                ->label('Modo de Capitalização')
                                ->options([
                                    'SIMPLES' => 'Simples (Linear)',
                                    'COMPOSTO' => 'Composto (Exponencial)',
                                ])
                                ->default('COMPOSTO')
                                ->visible(fn ($get) => $get('juros_tipo') === 'CONVENCIONAL_MENSAL'),
                            Toggle::make('juros_prorata')
                                ->label('Pro Rata Die')
                                ->default(true),
                        ]),
                        Grid::make(4)->schema([
                            Toggle::make('aplicar_art_523_multa')
                                ->label('Multa de 10% (Art. 523 CPC)')
                                ->default(true),
                            Toggle::make('aplicar_art_523_honorarios')
                                ->label('Honorários de 10% (Art. 523 CPC)')
                                ->default(true),
                            TextInput::make('multa_moratoria_percentual')
                                ->label('Multa Moratória/Contratual (%)')
                                ->numeric()
                                ->default(0.0),
                            TextInput::make('honorarios_sucumbenciais_percentual')
                                ->label('Honorários Sucumbenciais (%)')
                                ->numeric()
                                ->default(0.0),
                        ]),
                    ]),

                Section::make('Parcelas, Custas e Abatimentos')
                    ->description('Adicione débitos, despesas processuais ou pagamentos parciais.')
                    ->schema([
                        Repeater::make('parcelas')
                            ->label('Lista de Lançamentos')
                            ->schema([
                                TextInput::make('descricao')
                                    ->label('Descrição')
                                    ->required()
                                    ->placeholder('Ex: Parcela 01, Dano Material, Custas...'),
                                Select::make('tipo')
                                    ->label('Natureza')
                                    ->options([
                                        'DEBITO' => 'Débito Principal (sofre juros e multa)',
                                        'CUSTA_PROCESSUAL' => 'Custa Judicial (apenas correção)',
                                        'DESPESA_PROCESSUAL' => 'Despesa Processual (apenas correção)',
                                        'PAGAMENTO_PARCIAL' => 'Pagamento Parcial / Abatimento',
                                    ])
                                    ->default('DEBITO')
                                    ->required(),
                                DatePicker::make('data_vencimento')
                                    ->label('Data de Vencimento')
                                    ->required(),
                                TextInput::make('valor_original')
                                    ->label('Valor Nominal (R$)')
                                    ->numeric()
                                    ->required()
                                    ->prefix('R$'),
                            ])
                            ->columns(4)
                            ->defaultItems(1)
                            ->addActionLabel('Adicionar Nova Parcela'),
                    ]),
            ])
            ->statePath('data');
    }

    public function calcular(CalculoJudicialService $service): void
    {
        $dadosForm = $this->form->getState();

        $payload = [
            'numero_processo' => $dadosForm['numero_processo'] ?? '',
            'data_atualizacao' => $dadosForm['data_atualizacao'],
            'indexador_codigo' => $dadosForm['indexador_codigo'],
            'aplicar_deflacao' => (bool)($dadosForm['aplicar_deflacao'] ?? false),
            'aplicar_prorata_correcao' => (bool)($dadosForm['aplicar_prorata_correcao'] ?? false),
            'imputar_art_354_cc' => (bool)($dadosForm['imputar_art_354_cc'] ?? false),
            'juros_mora' => [
                'tipo' => $dadosForm['juros_tipo'],
                'percentual_mensal' => (float)($dadosForm['juros_percentual'] ?? 1.0),
                'modo' => $dadosForm['juros_modo'] ?? 'SIMPLES',
                'aplicar_prorata_die' => (bool)($dadosForm['juros_prorata'] ?? true),
            ],
            'fase_execucao' => [
                'aplicar_art_523_multa' => (bool)($dadosForm['aplicar_art_523_multa'] ?? false),
                'aplicar_art_523_honorarios' => (bool)($dadosForm['aplicar_art_523_honorarios'] ?? false),
                'multa_moratoria_percentual' => (float)($dadosForm['multa_moratoria_percentual'] ?? 0.0),
                'honorarios_sucumbenciais_percentual' => (float)($dadosForm['honorarios_sucumbenciais_percentual'] ?? 0.0),
            ],
            'parcelas' => array_map(function ($item, $idx) {
                return [
                    'numero_item' => $idx + 1,
                    'descricao' => $item['descricao'] ?? "Item " . ($idx + 1),
                    'tipo' => $item['tipo'],
                    'data_vencimento' => $item['data_vencimento'],
                    'valor_original' => (float)$item['valor_original'],
                ];
            }, $dadosForm['parcelas'] ?? [], array_keys($dadosForm['parcelas'] ?? [])),
        ];

        $resp = $service->calcular($payload);

        if (!empty($resp['sucesso'])) {
            $this->resultado = $resp;
            Notification::make()
                ->title('Cálculo Concluído!')
                ->body('Liquidação judicial processada com sucesso.')
                ->success()
                ->send();
        } else {
            Notification::make()
                ->title('Erro ao Processar')
                ->body($resp['mensagem'] ?? 'Erro desconhecido ao comunicar com motor de cálculo.')
                ->danger()
                ->send();
        }
    }

    public function limpar(): void
    {
        $this->resultado = null;
        $this->dadosProcesso = null;
        $this->mount();
        Notification::make()->title('Formulário redefinido.')->info()->send();
    }

    public function emitirLaudoPdf(CalculoJudicialService $service)
    {
        if (empty($this->resultado)) {
            Notification::make()->title('Execute o cálculo primeiro.')->warning()->send();
            return null;
        }

        $dadosForm = $this->form->getState();
        $user = auth()->user() ?? auth()->guard('cliente')->user();
        $calculo = $service->salvarCalculoComLead(
            $dadosForm,
            $this->resultado,
            [
                'nome' => $user->name ?? $user->nome ?? 'Operador do Painel',
                'email' => $user->email ?? '',
                'oab' => 'Painel Oficial',
            ]
        );

        return redirect()->route('laudo.imprimir', $calculo->uuid);
    }

    public function exportarExcel(CalculoJudicialService $service)
    {
        if (empty($this->resultado)) {
            Notification::make()->title('Execute o cálculo primeiro.')->warning()->send();
            return null;
        }

        $conteudo = $service->exportarPlanilha($this->resultado, 'xlsx');
        if (!$conteudo) {
            Notification::make()->title('Erro ao gerar planilha Excel.')->danger()->send();
            return null;
        }

        return response()->streamDownload(function () use ($conteudo) {
            echo $conteudo;
        }, 'memoria-calculo-judicial.xlsx', [
            'Content-Type' => 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        ]);
    }

    protected function getHeaderActions(): array
    {
        return [
            Action::make('importarPlanilha')
                ->label('Importar Planilha / Colar')
                ->icon('heroicon-o-arrow-up-tray')
                ->color('success')
                ->form([
                    FileUpload::make('arquivo')
                        ->label('Arquivo Excel (.xlsx) ou CSV (.csv)')
                        ->acceptedFileTypes([
                            'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
                            'text/csv',
                            'text/plain',
                            'application/vnd.ms-excel'
                        ])
                        ->disk('local')
                        ->directory('temp_import'),
                    Textarea::make('texto_colado')
                        ->label('Ou cole as linhas copiadas do Excel (Ctrl+V)')
                        ->placeholder("15/01/2023\tParcela 01\t1500,00\n15/02/2023\tParcela 02\t1500,00")
                        ->rows(6),
                ])
                ->action(function (array $data, CalculoJudicialService $service) {
                    $resultadoImportacao = null;

                    if (!empty($data['arquivo'])) {
                        $path = Storage::disk('local')->path($data['arquivo']);
                        $resultadoImportacao = $service->importarPlanilha($path);
                        @unlink($path);
                    } elseif (!empty($data['texto_colado'])) {
                        $resultadoImportacao = $service->importarPlanilha($data['texto_colado']);
                    }

                    if ($resultadoImportacao && !empty($resultadoImportacao['ok']) && !empty($resultadoImportacao['parcelas'])) {
                        $novasParcelas = [];
                        foreach ($resultadoImportacao['parcelas'] as $p) {
                            $novasParcelas[] = [
                                'descricao' => $p['descricao'] ?? 'Parcela Importada',
                                'tipo' => $p['tipo'] ?? 'DEBITO',
                                'data_vencimento' => $p['data_vencimento'] ?? $p['data'] ?? now()->format('Y-m-d'),
                                'valor_original' => (float)($p['valor_original'] ?? 0),
                            ];
                        }
                        
                        $estadoAtual = $this->form->getState();
                        $estadoAtual['parcelas'] = $novasParcelas;
                        $this->form->fill($estadoAtual);

                        Notification::make()
                            ->title('Planilha Importada!')
                            ->body(count($novasParcelas) . ' parcela(s) carregada(s) com sucesso na tabela.')
                            ->success()
                            ->send();
                    } else {
                        Notification::make()
                            ->title('Falha na Importação')
                            ->body($resultadoImportacao['mensagem'] ?? 'Nenhum dado válido encontrado para importar.')
                            ->danger()
                            ->send();
                    }
                }),

            Action::make('exportarExcel')
                ->label('Exportar Excel (.xlsx)')
                ->icon('heroicon-o-table-cells')
                ->color('info')
                ->visible(fn () => !empty($this->resultado))
                ->action(fn (CalculoJudicialService $service) => $this->exportarExcel($service)),

            Action::make('emitirLaudo')
                ->label('Emitir Laudo Certificado (PDF)')
                ->icon('heroicon-o-document-check')
                ->color('primary')
                ->visible(fn () => !empty($this->resultado))
                ->action(fn (CalculoJudicialService $service) => $this->emitirLaudoPdf($service)),
        ];
    }
}
