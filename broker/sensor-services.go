package main

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

// listenUDPSensors escuta os pacotes UDP enviados pelos sensores e envia para os clientes.
// Se o sensor for novo, é registrado no map e logado. Se já existir, é atualizado o estado e a última vez que foi visto.
// Params:
// conn - a conexão UDP para escutar os pacotes dos sensores
func listenUDPSensors(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("erro ao ler UDP %v", err)
			continue
		}

		var packet structs.Sensor
		if err := json.Unmarshal(buffer[:n], &packet); err != nil {
			log.Printf("pacote sensor inválido: %v", err)
			continue
		}

		sensorState := SensorState{
			Sensor:   packet,
			LastSeen: time.Now(),
		}

		sensorsMapMutex.Lock()
		_, exists := sensorsMap[packet.SensorID]
		sensorsMap[packet.SensorID] = sensorState
		sensorsMapMutex.Unlock()

		if !exists {
			log.Printf("Novo sensor detectado: %s", packet.SensorID)
		}
		broadcastSensorToClients(sensorState)
	}
}

// monitorSensorTTL monitora e remove sensores que passaram do TTL. Varre a cada 5 segundos.
func monitorSensorTTL() {
	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {
		expiredIDs := make([]string, 0)
		sensorsMapMutex.RLock()

		for id, state := range sensorsMap {
			if time.Since(state.LastSeen) > TTLSENSOR {
				log.Printf("Sensor %s marcado para expiração", id)
				expiredIDs = append(expiredIDs, id)
			}
		}

		sensorsMapMutex.RUnlock()

		if len(expiredIDs) == 0 {
			continue
		}

		for _, id := range expiredIDs {
			sensorsMapMutex.Lock()
			sensorState := sensorsMap[id]

			if time.Since(sensorState.LastSeen) > TTLSENSOR {
				delete(sensorsMap, id)
				log.Printf("Sensor %s expirado e removido", id)

				notifyRemovalToClients(structs.RemoveMessage{
					Type:   "remove",
					Entity: "sensor",
					ID:     id,
					Reason: "expirado",
				})
			}

			sensorsMapMutex.Unlock()

		}
	}
}
