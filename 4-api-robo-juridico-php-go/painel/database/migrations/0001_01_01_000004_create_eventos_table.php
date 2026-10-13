<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (!Schema::hasTable('eventos')) {
            Schema::create('eventos', function (Blueprint $table) {
                $table->id();
                $table->string('momento');
                $table->foreignId('admin_id')->nullable()->constrained('admins')->nullOnDelete();
                $table->string('admin_email')->default('');
                $table->string('acao');
                $table->string('alvo')->default('');
                $table->text('detalhe')->default('');
                $table->string('ip')->default('');

                $table->index('momento');
            });
        }
    }

    public function down(): void
    {
        Schema::dropIfExists('eventos');
    }
};
