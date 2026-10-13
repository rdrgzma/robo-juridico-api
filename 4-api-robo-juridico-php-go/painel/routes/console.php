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
