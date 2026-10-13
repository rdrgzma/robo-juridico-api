<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (!Schema::hasTable('admins')) {
            Schema::create('admins', function (Blueprint $table) {
                $table->id();
                $table->string('email')->unique();
                $table->string('nome')->default('');
                $table->string('senha_hash');
                $table->boolean('ativo')->default(true);
                $table->string('criado_em');
                $table->string('ultimo_login')->nullable();
                $table->rememberToken();
            });
        }
    }

    public function down(): void
    {
        Schema::dropIfExists('admins');
    }
};
