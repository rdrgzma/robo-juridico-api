<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('calculos_salvos', function (Blueprint $table) {
            $table->unsignedBigInteger('cliente_id')->nullable()->after('id')->index();
            $table->string('status_lead', 30)->default('novo')->after('lead_oab')->index();
            $table->text('observacoes_crm')->nullable()->after('status_lead');
        });
    }

    public function down(): void
    {
        Schema::table('calculos_salvos', function (Blueprint $table) {
            $table->dropColumn(['cliente_id', 'status_lead', 'observacoes_crm']);
        });
    }
};
