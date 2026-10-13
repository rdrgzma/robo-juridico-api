<?php

namespace App\Models;

use Carbon\Carbon;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class Uso extends Model
{
    use HasFactory;

    protected $table = 'uso';
    public $timestamps = false;

    protected $fillable = [
        'cliente_id',
        'chave_id',
        'momento',
        'metodo',
        'rota',
        'caminho',
        'status',
        'ms',
        'ip',
        'cobravel',
        'cobrado',
    ];

    protected function casts(): array
    {
        return [
            'status' => 'integer',
            'ms' => 'integer',
            'cobravel' => 'boolean',
            'cobrado' => 'boolean',
        ];
    }

    public function cliente(): BelongsTo
    {
        return $this->belongsTo(Cliente::class, 'cliente_id');
    }

    public function chave(): BelongsTo
    {
        return $this->belongsTo(Chave::class, 'chave_id');
    }

    /**
     * Retorna a data/hora formatada no horário oficial de Brasília.
     */
    public function getMomentoBrasiliaAttribute(): string
    {
        try {
            return Carbon::parse($this->momento)
                ->setTimezone('America/Sao_Paulo')
                ->format('d/m/Y H:i:s');
        } catch (\Throwable) {
            return $this->momento;
        }
    }
}
