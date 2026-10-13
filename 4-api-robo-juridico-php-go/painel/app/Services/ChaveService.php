<?php

namespace App\Services;

use App\Models\Chave;
use App\Models\Cliente;
use Illuminate\Support\Str;

class ChaveService
{
    /**
     * Gera um novo token seguro no padrão 'rj_live_<32_chars_base64url>'.
     * Salva apenas os primeiros 14 caracteres como prefixo e o SHA-256 no banco.
     * Devolve o token em texto puro UMA ÚNICA VEZ para exibição ao administrador.
     */
    public function gerarParaCliente(Cliente $cliente, string $nome = 'Chave Principal'): array
    {
        // 32 bytes aleatórios em base64url ~ 43 caracteres
        $random = rtrim(strtr(base64_encode(random_bytes(32)), '+/', '-_'), '=');
        $token = 'rj_live_' . $random;

        $prefixo = substr($token, 0, 16) . '...';
        $hashHex = hash('sha256', $token);

        $chave = Chave::create([
            'cliente_id' => $cliente->id,
            'nome' => $nome,
            'prefixo' => $prefixo,
            'hash' => $hashHex,
            'criada_em' => now()->utc()->toIso8601String(),
        ]);

        return [
            'token' => $token,
            'chave' => $chave,
        ];
    }
}
