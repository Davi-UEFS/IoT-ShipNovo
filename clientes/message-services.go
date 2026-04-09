package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

// receiveBrokerMessages lê mensagens do broker em loop, tratando cada tipo de mensagem (sensor, atuador, remoção, ACK).
// Params:
// reader - O leitor da conexão com o broker, para ler as mensagens.
func receiveBrokerMessages(reader *bufio.Reader, brokerAddr, clientID string) {

	for {
		conn := getBrokerConn()
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		line, err := reader.ReadBytes('\n')
		if err != nil {
			clearScreen()
			setOfflineStatus(true)
			functions.SafePrintf("Conexão com broker perdida: %v\n", err)
			publishUIEvent("Conexão com broker perdida. Tentando reconectar...")

			time.Sleep(time.Second) //WARNING: RECONEXAO ACONTECE ANTES DO BROKER FECHAR.HARDCODED
			newConn := functions.ConnectWithRetry(brokerAddr)
			if newConn == nil {
				functions.SafeFatalf("Não foi possível estabelecer conexão com o broker em %s", brokerAddr)
			}

			reader = bufio.NewReader(newConn)

			if err := functions.DoHandshake(newConn, reader, clientID, "client"); err != nil {
				functions.SafeFatalf("Erro no handshake após reconexão: %v", err)
			}
			functions.SafePrintf("Handshake OK após reconexão (actuator id=%s)\n", clientID)

			setBrokerConn(newConn)
			setOfflineStatus(false)
			publishUIEvent(fmt.Sprintf("Você se reconectou ao broker (%v)", time.Now().Format("15:04:05")))

			continue
		}

		var env structs.EnvelopeMessage // EnvelopeMessage é uma struct genérica para identificar o tipo da mensagem antes de fazer o unmarshal específico.

		if err := json.Unmarshal(line, &env); err != nil {
			functions.SafePrintf("Mensagem com formato inválido do broker: %v\n", err)
			continue
		}

		switch env.Type {
		case "ping":
			conn := getBrokerConn()
			if conn != nil {
				if err := functions.WriteJSONLine(conn, structs.PingPongMessage{Type: "pong"}); err != nil {
					functions.SafePrintf("Erro ao responder ping do broker: %v\n", err)
				}
			}

		case "sensor":

			var pkt structs.Sensor
			if err := json.Unmarshal(line, &pkt); err != nil {
				continue
			}

			sensorsMapMutex.Lock()
			_, exists := sensorsMap[pkt.SensorID]
			sensorsMap[pkt.SensorID] = pkt
			sensorsMapMutex.Unlock()

			if !exists {
				publishUIEvent("Novo sensor adicionado: " + pkt.SensorID)
			}

		case "actuator":
			var pkt structs.Actuator
			if err := json.Unmarshal(line, &pkt); err != nil {
				continue
			}
			actuatorsMapMutex.Lock()
			_, exists := actuatorsMap[pkt.ID]
			actuatorsMap[pkt.ID] = pkt
			actuatorsMapMutex.Unlock()

			if !exists {
				publishUIEvent("Novo atuador adicionado: " + pkt.ID)
			}

		case "remove":
			var remMsg structs.RemoveMessage
			if err := json.Unmarshal(line, &remMsg); err != nil {
				functions.SafePrintf("Mensagem de remoção do broker com formato inválido: %s | raw=%s\n", remMsg.ID, string(line))
				continue
			}

			removeEntity(remMsg)

		case "ack":
			var ackMsg structs.ACKMessage
			if err := json.Unmarshal(line, &ackMsg); err != nil {
				functions.SafePrintf("ACK do broker com formato inválido: %v | raw=%s\n", err, string(line))
				continue
			}
			deliverAck(ackMsg.RequestID, PendingAck{ackMsg.OK, ackMsg.Message})

		default:
			functions.SafePrintf("Tipo desconhecido do broker: %q | raw=%s\n", env.Type, string(line))
		}
	}
}

// removeEntity remove um sensor ou atuador do map.
// Anuncia a remoção como um evento de UI
// Params: message - A mensagem de remoção contendo informações sobre a entidade a ser removida.
func removeEntity(message structs.RemoveMessage) {
	switch message.Entity {
	case "actuator":
		actuatorsMapMutex.Lock()
		delete(actuatorsMap, message.ID)
		actuatorsMapMutex.Unlock()
		publishUIEvent("O atuador " + message.ID + " foi removido do sistema. Motivo: " + message.Reason)

	case "sensor":
		sensorsMapMutex.Lock()
		delete(sensorsMap, message.ID)
		sensorsMapMutex.Unlock()
		publishUIEvent("O sensor " + message.ID + " foi removido do sistema. Motivo: " + message.Reason)
	default:
		functions.SafePrintf("Mensagem de remoção com entidade desconhecida do broker: %s | raw=%s\n", message.Entity, string(message.ID))
	}
}

// publishUIEvent coloca uma mensagem no canal de notificações de UI para ser exibida ao usuário.
// Params:
// message - A mensagem de evento a ser exibida na UI.
func publishUIEvent(message string) {
	eventsChannel <- UIEvents{
		Message: message,
	}
}
