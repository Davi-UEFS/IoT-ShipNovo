package main

import "time"

// ── Configuração por tipo ──
// Cada tipo de sensor tem um intervalo diferente para geração de dados.
var timeIntervals = map[string]time.Duration{
	"temperatura": 500 * time.Millisecond,
	"combustivel": 1000 * time.Millisecond,
	"porcao":      1000 * time.Millisecond,
}

// Tipos de sensores disponíveis no sistema.
var sensorTypes = []string{"temperatura", "combustivel", "porcao"}