<?php

namespace App\Models;

use Carbon\Carbon;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class Evento extends Model
{
    use HasFactory;

    protected $table = 'eventos';
    public $timestamps = false;

    protected $fillable = [
        'momento',
        'admin_id',
        'admin_email',
        'acao',
        'alvo',
        'detalhe',
        'ip',
    ];

    public function admin(): BelongsTo
    {
        return $this->belongsTo(Admin::class, 'admin_id');
    }

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
