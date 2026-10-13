<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;

class Chave extends Model
{
    use HasFactory;

    protected $table = 'chaves';
    public $timestamps = false;

    protected $fillable = [
        'cliente_id',
        'nome',
        'prefixo',
        'hash',
        'criada_em',
        'revogada_em',
        'ultimo_uso',
    ];

    public function cliente(): BelongsTo
    {
        return $this->belongsTo(Cliente::class, 'cliente_id');
    }

    public function usos(): HasMany
    {
        return $this->hasMany(Uso::class, 'chave_id');
    }

    public function isAtiva(): bool
    {
        return is_null($this->revogada_em);
    }

    public function revogar(): void
    {
        $this->update([
            'revogada_em' => now()->utc()->toIso8601String(),
        ]);
    }
}
