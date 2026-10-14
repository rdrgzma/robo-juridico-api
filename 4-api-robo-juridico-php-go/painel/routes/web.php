<?php

use App\Http\Controllers\PublicCalculadoraController;
use Illuminate\Support\Facades\Route;

// Redireciona raiz ou serve a Calculadora Pública Lead Magnet
Route::get('/', [PublicCalculadoraController::class, 'index'])->name('home');
Route::get('/calculadora', [PublicCalculadoraController::class, 'index'])->name('calculadora.publica');

// Endpoints AJAX para a Calculadora Pública
Route::post('/calculadora/calcular', [PublicCalculadoraController::class, 'calcular'])->name('calculadora.calcular');
Route::post('/calculadora/importar-planilha', [PublicCalculadoraController::class, 'importarPlanilha'])->name('calculadora.importar-planilha');
Route::post('/calculadora/emitir-laudo', [PublicCalculadoraController::class, 'emitirLaudo'])->name('calculadora.emitir-laudo');

// Emissão e Impressão de Laudo Pericial Certificado (com QR Code)
Route::get('/laudo/{uuid}/imprimir', [PublicCalculadoraController::class, 'imprimirLaudo'])->name('laudo.imprimir');

// Validação Pública de Autenticidade (Link do QR Code)
Route::get('/validar/{uuid}', [PublicCalculadoraController::class, 'validar'])->name('validar.calculo');
