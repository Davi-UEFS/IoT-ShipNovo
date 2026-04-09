package main

import (
	"math/rand"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
)

// ── Funções de geração ──

// generateTemp varia a temperatura do motor com pequenas flutuações, simulando operação normal.
// Params:
// base 	- A temperatura anterior do motor, que é modificada pela função para criar variações realistas
// Returns:
// float64 	- A temperatura atualizada do motor.
func generateTemp(base *float64) float64 {
	variance := (rand.Float64() - 0.5) * 2 // variação entre -1 e +1
	*base += variance
	return *base
}

// generateGasVolume decrementa gradualmente o nível do tanque.
// Reinicia ao chegar em 0%.
// Params:
// volume 	- O nível atual do tanque, que é reduzido pela função para simular consumo de combustível
// Returns:
// float64 	- O novo nível do tanque.
func generateGasVolume(volume *float64) float64 {
	*volume -= rand.Float64() * 0.3 // consome até 0.3% por leitura
	if *volume < 0 {
		*volume = 100.0 // reinicia o tanque (simulação)
		functions.SafePrintln(" Tanque reabastecido — reiniciando para 100")
	}
	return *volume
}

// generateWaterLevel sobe gradualmente simulando acúmulo de água no porão.
// Reinicia ao atingir 100cm para a demo continuar.
// Params:
// level 	- O nível atual da água, que é aumentado pela função para simular acúmulo de água no porão
// Returns:
// float64 	- O novo nível da água.
func generateWaterLevel(level *float64) float64 {
	*level += rand.Float64() * 0.5 // acumula até 0.5cm por leitura
	if *level > 100 {
		*level = 0.0
		functions.SafePrintln("  Porão esvaziado — reiniciando para 0cm")
	}
	return *level
}
