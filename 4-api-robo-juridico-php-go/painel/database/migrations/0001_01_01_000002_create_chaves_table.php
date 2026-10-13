<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (!Schema::hasTable('chaves')) {
            Schema::create('chaves', function (Blueprint $table) {
                $table->id();
                $table->foreignId('cliente_id')->constrained('clientes')->cascadeOnDelete();
                $table->string('nome')->default('');
                $table->string('prefixo');
                $table->string('hash')->unique();
                $table->string('criada_em');
                $table->string('revogada_em')->nullable();
                $table->string('ultimo_uso')->nullable();
            });
        }
    }

    public function down(): void
    {
        Schema::dropIfExists('chaves');
    }
};
