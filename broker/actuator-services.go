package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

// registerActuator registra a conexão de um atuador no map do broker.
// Utiliza placeholder pois o atuador envia seu estado após handshake.
// Params:
//
//	conn 		- A conexão
//	reader 		- O leitor da conexão
//	actuatorID 	- O ID do atuador
func registerActuator(conn net.Conn, reader *bufio.Reader, actuatorID string) {

	actuatorsMapMutex.Lock()
	actuatorsMap[actuatorID] = &Actuator{
		ID:         actuatorID,
		State:      "placeholder",
		LastAction: "placeholder",
		Conn:       conn,
		LastPong:   time.Now(),
	}

	log.Printf("Atuador registrado: %s (%s)", actuatorID, conn.RemoteAddr().String())

	go pingActuator(actuatorsMap[actuatorID])
	actuatorsMapMutex.Unlock()
	handleActuatorConnection(reader, actuatorID)
}

// handleActuatorConnection trata a conexão com o atuador.
// Params:
//
//	reader 		- O leitor da conexão
//	actuatorID 	- O ID do atuador
func handleActuatorConnection(reader *bufio.Reader, actuatorID string) {
	for {

		line, err := reader.ReadBytes('\n')
		if err != nil {
			log.Printf("Atuador desconectou: %s", actuatorID)
			actuatorsMapMutex.Lock()
			delete(actuatorsMap, actuatorID)
			actuatorsMapMutex.Unlock()

			_ = notifyRemovalToClients(structs.RemoveMessage{
				Type:   "remove",
				Entity: "actuator",
				ID:     actuatorID,
				Reason: "desconectado",
			})

			return
		}

		var env structs.EnvelopeMessage
		if err := json.Unmarshal(line, &env); err != nil {
			log.Printf("Mensagem inválida do atuador %s: %v", actuatorID, err)
			continue
		}

		if env.Type == "pong" {
			actuatorsMapMutex.Lock()
			if act, ok := actuatorsMap[actuatorID]; ok && act.Conn != nil {
				act.LastPong = time.Now()
			}
			actuatorsMapMutex.Unlock()
			continue
		}

		var actPkg Actuator
		if err := json.Unmarshal(line, &actPkg); err != nil {
			continue
		}

		actuatorsMapMutex.Lock()
		if act, ok := actuatorsMap[actuatorID]; ok {
			// usa actuatorID do handshake como referência principal
			act.State = actPkg.State
			act.LastAction = actPkg.LastAction
		}
		actuatorsMapMutex.Unlock()

		log.Printf("\nO atuador = %s foi atualizado! \nEstado = %s \nÚltima ação = %s",
			actuatorID, actPkg.State, actPkg.LastAction)

		pendingActuatorsUpdate.Done()
	}
}

// sendCommandToActuator envia o comando do cliente para o atuador.
// Params:
// cmd		- O comando
// Returns:
// erro		- Se o comando falhar
func sendCommandToActuator(cmd structs.ClientCommand) error {
	if cmd.Role != "client" || cmd.ActuatorID == "" || cmd.Action == "" {
		return fmt.Errorf("Comando inválido")
	}

	actuatorsMapMutex.RLock()
	act, ok := actuatorsMap[cmd.ActuatorID]
	actuatorsMapMutex.RUnlock()

	if !ok || act == nil || act.Conn == nil {
		return fmt.Errorf("Atuador %s não encontrado/conectado", cmd.ActuatorID)
	}

	act.writeMutex.Lock()
	defer act.writeMutex.Unlock()
	return functions.WriteJSONLine(act.Conn, structs.ActuatorCommand{
		Action: cmd.Action,
	})
}

func updateActuators() {
	actuatorsMapMutex.RLock()
	defer actuatorsMapMutex.RUnlock()

	for _, act := range actuatorsMap {
		pendingActuatorsUpdate.Add(1)
		log.Printf("Solicitando atualização do atuador %s...", act.ID)
		act.writeMutex.Lock()
		functions.WriteJSONLine(act.Conn, structs.ActuatorCommand{
			Action: "status",
		})
		act.writeMutex.Unlock()
	}

}

func pingActuator(act *Actuator) {

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		actuatorsMapMutex.RLock()
		currentAct, ok := actuatorsMap[act.ID]
		if !ok || currentAct != act {
			actuatorsMapMutex.RUnlock()
			return
		}
		lastPong := act.LastPong
		actuatorsMapMutex.RUnlock()

		if time.Since(lastPong) > 4*time.Second {
			log.Printf("Atuador %s sem pong dentro do prazo; encerrando conexão", act.ID)
			_ = act.Conn.Close()
			return
		}

		act.writeMutex.Lock()
		err := functions.WriteJSONLine(act.Conn, structs.PingPongMessage{Type: "ping"})
		act.writeMutex.Unlock()

		if err != nil {
			log.Printf("Erro ao enviar ping para atuador %s: %v", act.ID, err)
			_ = act.Conn.Close()
			return
		}
	}
}
