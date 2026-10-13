<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (Schema::hasTable('clientes')) {
            Schema::table('clientes', function (Blueprint $table) {
                if (!Schema::hasColumn('clientes', 'senha_hash')) {
                    $table->string('senha_hash')->nullable()->after('email');
                }
                if (!Schema::hasColumn('clientes', 'ultimo_login')) {
                    $table->string('ultimo_login')->nullable()->after('ativo');
                }
                if (!Schema::hasColumn('clientes', 'remember_token')) {
                    $table->rememberToken();
                }
            });
        }
    }

    public function down(): void
    {
        if (Schema::hasTable('clientes')) {
            Schema::table('clientes', function (Blueprint $table) {
                $table->dropColumn(['senha_hash', 'ultimo_login', 'remember_token']);
            });
        }
    }
};
