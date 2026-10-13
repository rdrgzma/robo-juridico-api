<?php

namespace App\Models;

use Carbon\Carbon;
use Filament\Models\Contracts\FilamentUser;
use Filament\Models\Contracts\HasName;
use Filament\Panel;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Foundation\Auth\User as Authenticatable;
use Illuminate\Notifications\Notifiable;

class Cliente extends Authenticatable implements FilamentUser, HasName
{
    use HasFactory, Notifiable;

    protected $table = 'clientes';
    public $timestamps = false;

    protected $fillable = [
        'nome',
        'email',
        'senha_hash',
        'empresa',
        'plano',
        'valor_mensal_centavos',
        'limite_mensal',
        'limite_por_minuto',
        'ativo',
        'observacoes',
        'ultimo_login',
        'criado_em',
        'atualizado_em',
    ];

    protected $hidden = [
        'senha_hash',
        'remember_token',
    ];

    protected function casts(): array
    {
        return [
            'ativo' => 'boolean',
            'valor_mensal_centavos' => 'integer',
            'limite_mensal' => 'integer',
            'limite_por_minuto' => 'integer',
        ];
    }

    public function getAuthPasswordName(): string
    {
        return 'senha_hash';
    }

    public function canAccessPanel(Panel $panel): bool
    {
        if ($panel->getId() === 'portal') {
            return (bool) $this->ativo && !empty($this->senha_hash);
        }

        return false;
    }

    public function getFilamentName(): string
    {
        return $this->nome ?: ($this->empresa ?: $this->email);
    }

    public function chaves(): HasMany
    {
        return $this->hasMany(Chave::class, 'cliente_id');
    }

    public function chavesAtivas(): HasMany
    {
        return $this->hasMany(Chave::class, 'cliente_id')->whereNull('revogada_em');
    }

    public function usos(): HasMany
    {
        return $this->hasMany(Uso::class, 'cliente_id');
    }

    /**
     * Calcula o total de consultas faturadas consumidas no mês corrente (horário de Brasília).
     */
    public function consumoMesAtual(): int
    {
        $inicioMes = Carbon::now('America/Sao_Paulo')->startOfMonth()->utc()->toIso8601String();

        return (int) $this->usos()
            ->where('momento', '>=', $inicioMes)
            ->where('cobravel', 1)
            ->where(function ($q) {
                $q->whereNull('status')->orWhere('status', '<', 400);
            })
            ->count();
    }

    public function saldoRestante(): int
    {
        return max(0, $this->limite_mensal - $this->consumoMesAtual());
    }

    public function porcentagemConsumo(): float
    {
        if ($this->limite_mensal <= 0) {
            return 0.0;
        }
        return min(100.0, round(($this->consumoMesAtual() / $this->limite_mensal) * 100, 1));
    }
}
