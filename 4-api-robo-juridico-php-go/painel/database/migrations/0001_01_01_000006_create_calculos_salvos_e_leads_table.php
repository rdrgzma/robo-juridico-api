<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('calculos_salvos', function (Blueprint $table) {
            $table->id();
            $table->uuid('uuid')->unique();
            $table->string('hash_validacao', 64)->index();
            $table->string('numero_processo')->nullable()->index();
            $table->date('data_atualizacao');
            $table->string('indexador_codigo', 50);
            $table->decimal('total_geral', 14, 2);
            $table->json('payload_requisicao');
            $table->json('resultado_json');
            $table->string('lead_nome')->nullable();
            $table->string('lead_email')->nullable()->index();
            $table->string('lead_whatsapp')->nullable();
            $table->string('lead_oab')->nullable();
            $table->string('ip_origem', 45)->nullable();
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('calculos_salvos');
    }
};
