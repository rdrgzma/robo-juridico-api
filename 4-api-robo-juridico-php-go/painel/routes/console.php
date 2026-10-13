<?php

use Illuminate\Support\Facades\Artisan;

Artisan::command('admin:criar {email} {--nome=} {--senha=}', function ($email) {
    $nome = $this->option('nome') ?: explode('@', $email)[0];
    $senha = $this->option('senha') ?: 'admin123';

    $admin = \App\Models\Admin::updateOrCreate(
        ['email' => $email],
        [
            'nome' => $nome,
            'senha_hash' => bcrypt($senha),
            'ativo' => 1,
            'criado_em' => now()->toIso8601String(),
        ]
    );

    $this->info("Administrador '{$admin->email}' criado com sucesso! Senha: {$senha}");
})->purpose('Cria um novo administrador para o painel Filament');

Artisan::command('cliente:criar {email} {--nome=} {--senha=} {--empresa=} {--plano=pro}', function ($email) {
    $nome = $this->option('nome') ?: explode('@', $email)[0];
    $senha = $this->option('senha') ?: 'cliente123';
    $empresa = $this->option('empresa') ?: 'Empresa Parceira';
    $plano = $this->option('plano') ?: 'pro';
    $agora = now()->utc()->toIso8601String();

    $cliente = \App\Models\Cliente::updateOrCreate(
        ['email' => $email],
        [
            'nome' => $nome,
            'empresa' => $empresa,
            'plano' => $plano,
            'senha_hash' => bcrypt($senha),
            'limite_mensal' => 5000,
            'limite_por_minuto' => 60,
            'ativo' => 1,
            'criado_em' => $agora,
            'atualizado_em' => $agora,
        ]
    );

    $this->info("Cliente '{$cliente->email}' criado com sucesso para o Portal! Senha: {$senha}");
})->purpose('Cria ou atualiza um cliente com credenciais para o Portal do Desenvolvedor');
