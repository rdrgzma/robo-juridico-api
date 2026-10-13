<?php

namespace Tests\Feature;

use App\Models\Admin;
use App\Models\Cliente;
use Filament\Panel;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Hash;
use Tests\TestCase;

class ClientePortalAuthTest extends TestCase
{
    use RefreshDatabase;

    public function test_cliente_authenticates_with_guard_cliente(): void
    {
        $cliente = Cliente::create([
            'nome' => 'Advocacia Ramos',
            'email' => 'ramos@advocacia.com.br',
            'senha_hash' => Hash::make('senha123456'),
            'plano' => 'pro',
            'limite_mensal' => 5000,
            'limite_por_minuto' => 60,
            'ativo' => true,
            'criado_em' => now()->utc()->toIso8601String(),
            'atualizado_em' => now()->utc()->toIso8601String(),
        ]);

        $credenciaisValidas = [
            'email' => 'ramos@advocacia.com.br',
            'password' => 'senha123456',
        ];

        $this->assertTrue(Auth::guard('cliente')->attempt($credenciaisValidas));
        $this->assertEquals($cliente->id, Auth::guard('cliente')->id());

        $credenciaisInvalidas = [
            'email' => 'ramos@advocacia.com.br',
            'password' => 'senha-incorreta',
        ];

        $this->assertFalse(Auth::guard('cliente')->attempt($credenciaisInvalidas));
    }

    public function test_cliente_can_access_portal_panel_only(): void
    {
        $cliente = new Cliente([
            'email' => 'cliente@exemplo.com',
            'ativo' => true,
            'senha_hash' => Hash::make('senha123'),
        ]);

        $panelPortal = (new Panel())->id('portal');
        $panelAdmin = (new Panel())->id('admin');

        $this->assertTrue($cliente->canAccessPanel($panelPortal));
        $this->assertFalse($cliente->canAccessPanel($panelAdmin), 'Cliente não deve ter acesso ao painel admin');
    }

    public function test_inactive_or_passwordless_cliente_denied_portal_access(): void
    {
        $clienteInativo = new Cliente([
            'email' => 'inativo@exemplo.com',
            'ativo' => false,
            'senha_hash' => Hash::make('senha123'),
        ]);

        $clienteSemSenha = new Cliente([
            'email' => 'semsenha@exemplo.com',
            'ativo' => true,
            'senha_hash' => null,
        ]);

        $panelPortal = (new Panel())->id('portal');

        $this->assertFalse($clienteInativo->canAccessPanel($panelPortal));
        $this->assertFalse($clienteSemSenha->canAccessPanel($panelPortal));
    }

    public function test_admin_cannot_access_client_portal(): void
    {
        $admin = new Admin([
            'email' => 'admin@empresa.com.br',
            'ativo' => true,
            'senha_hash' => Hash::make('admin123'),
        ]);

        $panelPortal = (new Panel())->id('portal');
        $panelAdmin = (new Panel())->id('admin');

        $this->assertFalse($admin->canAccessPanel($panelPortal), 'Admin não deve ter acesso ao portal do cliente');
        $this->assertTrue($admin->canAccessPanel($panelAdmin));
    }
}
