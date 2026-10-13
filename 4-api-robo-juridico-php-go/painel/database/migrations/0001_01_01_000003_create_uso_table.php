<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (!Schema::hasTable('uso')) {
            Schema::create('uso', function (Blueprint $table) {
                $table->id();
                $table->foreignId('cliente_id')->nullable()->constrained('clientes')->cascadeOnDelete();
                $table->foreignId('chave_id')->nullable()->constrained('chaves')->nullOnDelete();
                $table->string('momento');
                $table->string('metodo')->default('GET');
                $table->string('rota')->default('');
                $table->string('caminho')->default('');
                $table->integer('status')->nullable();
                $table->integer('ms')->nullable();
                $table->string('ip')->default('');
                $table->boolean('cobravel')->default(false);
                $table->boolean('cobrado')->default(false);

                $table->index(['cliente_id', 'momento']);
                $table->index('momento');
            });
        }
    }

    public function down(): void
    {
        Schema::dropIfExists('uso');
    }
};
