<?php

namespace Database\Seeders;

use App\Models\Admin;
use Carbon\Carbon;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\Hash;

class AdminSeeder extends Seeder
{
    public function run(): void
    {
        Admin::firstOrCreate(
            ['email' => 'admin@empresa.com.br'],
            [
                'nome' => 'Administrador do Sistema',
                'senha_hash' => Hash::make('admin123'),
                'ativo' => true,
                'criado_em' => Carbon::now()->utc()->toIso8601String(),
            ]
        );
    }
}
