<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class CalculoSalvo extends Model
{
    protected $table = 'calculos_salvos';

    protected $fillable = [
        'cliente_id',
        'uuid',
        'hash_validacao',
        'numero_processo',
        'data_atualizacao',
        'indexador_codigo',
        'total_geral',
        'payload_requisicao',
        'resultado_json',
        'lead_nome',
        'lead_email',
        'lead_whatsapp',
        'lead_oab',
        'status_lead',
        'observacoes_crm',
        'ip_origem',
    ];

    protected $casts = [
        'payload_requisicao' => 'array',
        'resultado_json' => 'array',
        'total_geral' => 'decimal:2',
        'data_atualizacao' => 'date',
    ];

    public function cliente(): BelongsTo
    {
        return $this->belongsTo(Cliente::class, 'cliente_id');
    }
}
