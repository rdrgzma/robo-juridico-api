<?php

namespace Tests\Feature;

use App\Models\Admin;
use Filament\Panel;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Hash;
use Tests\TestCase;

class AdminAuthTest extends TestCase
{
    use RefreshDatabase;

    public function test_admin_model_authenticates_with_senha_hash_column(): void
    {
        $admin = Admin::create([
            'email' => 'teste@empresa.com.br',
            'nome' => 'Admin Teste',
            'senha_hash' => Hash::make('senha-secreta-123'),
            'ativo' => true,
            'criado_em' => now()->utc()->toIso8601String(),
        ]);

        $credenciaisValidas = [
            'email' => 'teste@empresa.com.br',
            'password' => 'senha-secreta-123',
        ];

        $this->assertTrue(Auth::guard('web')->attempt($credenciaisValidas));
        $this->assertEquals($admin->id, Auth::guard('web')->id());

        $credenciaisInvalidas = [
            'email' => 'teste@empresa.com.br',
            'password' => 'senha-errada',
        ];

        $this->assertFalse(Auth::guard('web')->attempt($credenciaisInvalidas));
    }

    public function test_active_admin_can_access_filament_panel(): void
    {
        $admin = new Admin([
            'email' => 'ativo@empresa.com.br',
            'ativo' => true,
        ]);

        $panel = new Panel();
        $this->assertTrue($admin->canAccessPanel($panel));
    }

    public function test_inactive_admin_is_denied_filament_panel_access(): void
    {
        $admin = new Admin([
            'email' => 'inativo@empresa.com.br',
            'ativo' => false,
        ]);

        $panel = new Panel();
        $this->assertFalse($admin->canAccessPanel($panel));
    }
}
