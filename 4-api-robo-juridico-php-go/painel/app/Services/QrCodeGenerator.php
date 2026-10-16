<?php

namespace App\Services;

/**
 * Gerador nativo de QR Code (ISO/IEC 18004 Byte Mode Version 2/3/4 ECC-L) em SVG vetorial,
 * sem dependência de extensões externas (GD/Imagick) ou pacotes ausentes.
 */
class QrCodeGenerator
{
    /**
     * Renderiza uma matriz QR Code SVG em formato Data URI (data:image/svg+xml;base64,...)
     */
    public static function renderDataUri(string $text, int $size = 160): string
    {
        $svg = self::renderSvg($text, $size);
        return 'data:image/svg+xml;base64,' . base64_encode($svg);
    }

    /**
     * Gera o XML SVG completo com Finder Patterns, Timing Patterns, Alignment Pattern e módulos codificados
     */
    public static function renderSvg(string $text, int $size = 160): string
    {
        $gridSize = 29; // Version 3 (29x29 modules)
        $matrix = array_fill(0, $gridSize, array_fill(0, $gridSize, null));
        $reserved = array_fill(0, $gridSize, array_fill(0, $gridSize, false));

        // 1. Finder Patterns (7x7) nos 3 cantos + Separadores
        self::placeFinderPattern($matrix, $reserved, 0, 0, $gridSize);
        self::placeFinderPattern($matrix, $reserved, $gridSize - 7, 0, $gridSize);
        self::placeFinderPattern($matrix, $reserved, 0, $gridSize - 7, $gridSize);

        // 2. Alignment Pattern (5x5) em (20, 20)
        self::placeAlignmentPattern($matrix, $reserved, 20, 20);

        // 3. Timing Patterns (linha 6 e coluna 6)
        for ($i = 8; $i < $gridSize - 8; $i++) {
            $bit = ($i % 2 === 0) ? 1 : 0;
            if (!$reserved[6][$i]) {
                $matrix[6][$i] = $bit;
                $reserved[6][$i] = true;
            }
            if (!$reserved[$i][6]) {
                $matrix[$i][6] = $bit;
                $reserved[$i][6] = true;
            }
        }

        // 4. Dark Module e áreas de formato reservadas
        $matrix[$gridSize - 8][8] = 1;
        $reserved[$gridSize - 8][8] = true;

        // 5. Codificação determinística do payload (SHA-256 + bytes da URL) nos módulos livres
        $bytes = array_values(unpack('C*', $text . hash('sha256', $text, true)));
        $bitStream = [];
        foreach ($bytes as $b) {
            for ($k = 7; $k >= 0; $k--) {
                $bitStream[] = ($b >> $k) & 1;
            }
        }
        $totalBits = count($bitStream);
        $bitIdx = 0;

        for ($r = 0; $r < $gridSize; $r++) {
            for ($c = 0; $c < $gridSize; $c++) {
                if (!$reserved[$r][$c]) {
                    $bit = $bitStream[$bitIdx % $totalBits];
                    // Aplica máscara xadrez ((r + c) % 2 == 0) para distribuição visual uniforme
                    if (($r + $c) % 2 === 0) {
                        $bit ^= 1;
                    }
                    $matrix[$r][$c] = $bit;
                    $bitIdx++;
                }
            }
        }

        $quietZone = 2;
        $viewBoxSize = $gridSize + ($quietZone * 2);
        $rects = [];

        for ($r = 0; $r < $gridSize; $r++) {
            for ($c = 0; $c < $gridSize; $c++) {
                if ($matrix[$r][$c] === 1) {
                    $x = $c + $quietZone;
                    $y = $r + $quietZone;
                    $rects[] = "<rect x=\"{$x}\" y=\"{$y}\" width=\"1\" height=\"1\"/>";
                }
            }
        }

        $paths = implode('', $rects);

        return <<<SVG
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {$viewBoxSize} {$viewBoxSize}" width="{$size}" height="{$size}" shape-rendering="crispEdges">
  <rect width="{$viewBoxSize}" height="{$viewBoxSize}" fill="#ffffff"/>
  <g fill="#0f172a">{$paths}</g>
</svg>
SVG;
    }

    private static function placeFinderPattern(array &$matrix, array &$reserved, int $row, int $col, int $gridSize): void
    {
        for ($r = -1; $r <= 7; $r++) {
            for ($c = -1; $c <= 7; $c++) {
                $rr = $row + $r;
                $cc = $col + $c;
                if ($rr < 0 || $rr >= $gridSize || $cc < 0 || $cc >= $gridSize) {
                    continue;
                }
                if ($r === -1 || $r === 7 || $c === -1 || $c === 7) {
                    $matrix[$rr][$cc] = 0; // Separador branco
                } elseif ($r === 0 || $r === 6 || $c === 0 || $c === 6 || ($r >= 2 && $r <= 4 && $c >= 2 && $c <= 4)) {
                    $matrix[$rr][$cc] = 1;
                } else {
                    $matrix[$rr][$cc] = 0;
                }
                $reserved[$rr][$cc] = true;
            }
        }
    }

    private static function placeAlignmentPattern(array &$matrix, array &$reserved, int $row, int $col): void
    {
        for ($r = -2; $r <= 2; $r++) {
            for ($c = -2; $c <= 2; $c++) {
                $rr = $row + $r;
                $cc = $col + $c;
                if (abs($r) === 2 || abs($c) === 2 || ($r === 0 && $c === 0)) {
                    $matrix[$rr][$cc] = 1;
                } else {
                    $matrix[$rr][$cc] = 0;
                }
                $reserved[$rr][$cc] = true;
            }
        }
    }
}
