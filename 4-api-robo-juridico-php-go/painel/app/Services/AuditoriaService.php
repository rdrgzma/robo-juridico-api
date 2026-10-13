<?php

namespace App\Services;

use App\Models\Evento;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Request;

class AuditoriaService
{
    public static function registrar(string $acao, string $alvo, string $detalhe = ''): Evento
    {
        $admin = Auth::user();

        return Evento::create([
            'momento' => now()->utc()->toIso8601String(),
            'admin_id' => $admin?->id,
            'admin_email' => $admin?->email ?? 'sistema',
            'acao' => $acao,
            'alvo' => $alvo,
            'detalhe' => $detalhe,
            'ip' => Request::ip() ?? '127.0.0.1',
        ]);
    }
}
