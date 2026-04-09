package main

import (
	"encoding/json"
	"net"
	"slices"
	"time"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

// checkSensorType checa se o tipo do sensor é válido, comparando com a lista de tipos permitidos.
// Params:
// tipo - O tipo do sensor a ser verificado (ex: "temperatura", "combustivel", "porcao").
// Returns:
// bool - Retorna true se o tipo for válido, ou false caso contrário.
func checkSensorType(tipo string) bool {
	return slices.Contains(sensorTypes, tipo)
}

// publishSensorLoop envia periodicamente os dados do sensor para o broker via UDP.
// Params: 
// conn 	- A conexão UDP com o broker.
// id 		- O ID do sensor, usado para identificação no broker.
// tipo 	- O tipo do sensor (ex: "temperatura", "combustivel", "porcao"), usado para gerar dados e identificação no broker.
func publishSensorLoop(conn *net.UDPConn, id string, tipo string) {
	// Valores base para geração de dados
	baseTemp := 85.0
	baseGas := 100.0
	baseWaterLevel := 0.0

	for {
		var valor float64
		switch tipo {
		case "temperatura":
			valor = generateTemp(&baseTemp)
		case "combustivel":
			valor = generateGasVolume(&baseGas)
		case "porcao":
			valor = generateWaterLevel(&baseWaterLevel)
		}

		tipoLabel := tipo
		switch tipo {
		case "combustivel":
			tipoLabel = "Combustível"
		case "temperatura":
			tipoLabel = "Temperatura"
		case "porcao":
			tipoLabel = "Porção"
		}

		functions.SafePrintf("Tipo: %s\nValor: %.2f\n", tipoLabel, valor)

		packet := structs.Sensor{
			SensorID: id,
			Type:     tipo,
			Value:    valor,
		}

		dados, err := json.Marshal(packet)
		if err != nil {
			functions.SafePrintf("[erro] falha ao serializar: %v\n", err)
			continue
		}

		conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))   // Tempo limite para escrita, evitando bloqueio indefinido
		_, err = conn.Write(dados)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				functions.SafePrintf("[aviso] timeout ao enviar UDP: %v\n", err)
			} else {
				functions.SafePrintf("[erro] falha ao enviar UDP: %v\n", err)
			}
		}

		time.Sleep(timeIntervals[tipo])   // Aguarda o intervalo específico do tipo de sensor antes de enviar o próximo dado
	}
}
