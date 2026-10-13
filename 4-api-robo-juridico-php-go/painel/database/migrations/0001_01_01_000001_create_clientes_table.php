<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (!Schema::hasTable('clientes')) {
            Schema::create('clientes', function (Blueprint $table) {
                $table->id();
                $table->string('nome');
                $table->string('email');
                $table->string('empresa')->default('');
                $table->string('plano')->default('basico');
                $table->integer('valor_mensal_centavos')->default(0);
                $table->integer('limite_mensal')->default(1000);
                $table->integer('limite_por_minuto')->default(60);
                $table->boolean('ativo')->default(true);
                $table->text('observacoes')->nullable();
                $table->string('criado_em');
                $table->string('atualizado_em');
            });
        }
    }

    public function down(): void
    {
        Schema::dropIfExists('clientes');
    }
};
