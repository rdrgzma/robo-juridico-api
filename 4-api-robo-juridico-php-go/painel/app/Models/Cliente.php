<?php

namespace App\Models;

use Carbon\Carbon;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;

class Cliente extends Model
{
    use HasFactory;

    protected $table = 'clientes';
    public $timestamps = false;

    protected $fillable = [
        'nome',
        'email',
        'empresa',
        'plano',
        'valor_mensal_centavos',
        'limite_mensal',
        'limite_por_minuto',
        'ativo',
        'observacoes',
        'criado_em',
        'atualizado_em',
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
