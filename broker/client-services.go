package main

import (
	"bufio"
	"encoding/json"
	"log"
	"maps"
	"net"
	"sync"
	"time"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

// registerClient registra a conexão de um cliente no map.
// Params:
//
//	conn 		- A conexão do cliente
//	reader 		- O leitor da conexão
//	clientID 	- O ID do cliente
func registerClient(conn net.Conn, reader *bufio.Reader, clientID string) {

	newSession := &ClientSession{
		ID:         clientID,
		Conn:       conn,
		writeMutex: sync.Mutex{},
		LastPong:   time.Now(),
	}

	clientsMapMutex.Lock()
	clientsSessionMap[clientID] = newSession
	clientsMapMutex.Unlock()

	log.Printf("Cliente registrado: %s (%s)", clientID, conn.RemoteAddr().String())

	if err := sendAllSensorsToNewClient(newSession); err != nil {
		log.Printf("Erro ao enviar snapshot inicial dos sensores para novo cliente %s: %v", clientID, err)
	}

	go pingClient(newSession)
	handleClientConnection(newSession, reader, clientID)
}

// handleClientConnection trata a conexão com o cliente.
// Params:
//
//	session		- A sessão do cliente, contendo conexão e mutex de escrita
//	reader 		- O leitor da conexão
//	clientID 	- O ID do cliente
func handleClientConnection(session *ClientSession, reader *bufio.Reader, clientID string) {
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {

			log.Printf("Erro de leitura do cliente %s: %v", clientID, err) // Se foi EOF ou outro erro
			clientsMapMutex.Lock()
			delete(clientsSessionMap, clientID)
			clientsMapMutex.Unlock()
			log.Printf("Cliente removido: %s", clientID)
			return
		}

		var env structs.EnvelopeMessage
		if err := json.Unmarshal(line, &env); err != nil {
			log.Printf("Mensagem inválida do cliente %s: %v", clientID, err)
			continue
		}

		if env.Type == "pong" {
			clientsMapMutex.Lock()
			if currentSession, ok := clientsSessionMap[clientID]; ok && currentSession == session {
				session.LastPong = time.Now()
			}
			clientsMapMutex.Unlock()
			continue
		}

		var cmd structs.ClientCommand
		if err := json.Unmarshal(line, &cmd); err != nil {
			log.Printf("Comando inválido de cliente %s: %v", clientID, err)
			continue
		}

		switch cmd.Type {
		case "command":

			if err := sendCommandToActuator(cmd); err != nil {

				log.Printf("Erro ao enviar comando para atuador %s: %v", cmd.ActuatorID, err)

				// Envia ACK de falha para o cliente poder sair do estado de espera
				if ackErr := sendACKToClient(session, structs.ACKMessage{
					Type:       "ack",
					RequestID:  cmd.RequestID,
					OK:         false,
					Message:    err.Error(),
					ActuatorID: cmd.ActuatorID,
					Action:     cmd.Action,
				}); ackErr != nil {
					log.Printf("Erro ao enviar ACK para o cliente %s \n", cmd.ClientID)
				}
				continue
			}

			log.Printf("Comando encaminhado: client = %s actuator = %s action = %s",
				clientID, cmd.ActuatorID, cmd.Action)

			// Envia ACK de sucesso ao cliente
			if ackErr := sendACKToClient(session, structs.ACKMessage{
				Type:       "ack",
				RequestID:  cmd.RequestID,
				OK:         true,
				ActuatorID: cmd.ActuatorID,
				Action:     cmd.Action,
			}); ackErr != nil {
				log.Printf("Erro ao enviar ACK de sucesso para cliente %s: %v", clientID, ackErr)
			}

		case "query":
			updateActuators()
			if err := sendActuatorsToClient(session); err != nil {
				log.Printf("Erro ao enviar atuadores para cliente %s: %v", clientID, err)
				if ackErr := sendACKToClient(session, structs.ACKMessage{
					Type:      "ack",
					RequestID: cmd.RequestID,
					OK:        false,
					Message:   err.Error(),
				}); ackErr != nil {
					log.Printf("Erro ao enviar ACK de falha para cliente %s: %v", clientID, ackErr)
				}
				continue
			}

			log.Printf("Atuadores enviados para cliente %s", clientID)

			if ackErr := sendACKToClient(session, structs.ACKMessage{
				Type:      "ack",
				RequestID: cmd.RequestID,
				OK:        true,
			}); ackErr != nil {
				log.Printf("Erro ao enviar ACK de sucesso para cliente %s: %v", clientID, ackErr)
			}

		default:
			log.Printf("Tipo de comando desconhecido do cliente %s: %q", clientID, cmd.Type)

		}
	}
}

// sendActuatorsToClient envia os atuadores registrados para o cliente.
// Params:
// session - A sessão do cliente para o qual os atuadores serão enviados
// Returns:
// error - Erro se algum envio apresentar falha
func sendActuatorsToClient(session *ClientSession) error {
	pendingActuatorsUpdate.Wait()

	actuatorsMapMutex.RLock()
	actuatorsCache := maps.Clone(actuatorsMap)
	actuatorsMapMutex.RUnlock()

	//TODO: PEDIR PARA O ATUADOR DIRETAMENTE ENVIAR O ESTADO ATUAL PARA O CLIENTE, AO INVÉS DE MANTER UM CACHE NO BROKER. ASSIM EVITA PROBLEMAS DE SINCRONIZAÇÃO E FALHAS DE ENVIO.

	session.writeMutex.Lock()
	defer session.writeMutex.Unlock()

	for _, act := range actuatorsCache {

		EnvelopeMessage := structs.ActuatorMessage{
			Type:       "actuator",
			ID:         act.ID,
			State:      act.State,
			LastAction: act.LastAction,
		}

		if err := functions.WriteJSONLine(session.Conn, EnvelopeMessage); err != nil {
			return err
		}

	}
	return nil
}

// sendACKToClient envia uma mensagem de ACK ao cliente para confirmar (ou rejeitar) um comando.
// Params:
// session - A sessão do cliente para o qual o ACK será enviado
// ack     - A mensagem de ACK a ser enviada
// Returns:
// error   - Erro se o envio do ACK apresentar falha
func sendACKToClient(session *ClientSession, ack structs.ACKMessage) error {
	session.writeMutex.Lock()
	defer session.writeMutex.Unlock()
	return functions.WriteJSONLine(session.Conn, ack)
}

// notifyRemovalToClients avisa a todos os clientes sobre a remoção de um atuador ou sensor.
// Params:
// message - A mensagem de remoção contendo o ID e tipo da entidade removida
// Returns:
// error   - Erro se algum envio apresentar falha
func notifyRemovalToClients(message structs.RemoveMessage) error {

	clientsMapMutex.RLock()
	clientsCache := maps.Clone(clientsSessionMap)
	clientsMapMutex.RUnlock()

	for clientID, session := range clientsCache {
		session.writeMutex.Lock()
		if err := functions.WriteJSONLine(session.Conn, message); err != nil {
			log.Printf("Erro ao notificar cliente %s sobre remoção: %v", clientID, err)
		}
		session.writeMutex.Unlock()
	}

	return nil
}

// broadcastSensorToClients envia a atualização de um sensor para todos os clientes conectados.
// Se o cliente estiver ocupado ou com muitas mensagens, a atualização é descartada para aliviar o fluxo.
// Params:
// sensorState - O sensor a ser enviado para os clientes
func broadcastSensorToClients(sensorState SensorState) {

	// Snapshot das conexões para evitar segurar o lock durante envios
	clientsMapMutex.RLock()
	clientsCache := maps.Clone(clientsSessionMap)
	clientsMapMutex.RUnlock()

	for clientID, session := range clientsCache {

		// Pula atualização se o cliente estiver ocupado/muito cheio.
		if session.writeMutex.TryLock() {

			EnvelopeMessage := structs.SensorMessage{
				Type:       "sensor",
				SensorID:   sensorState.Sensor.SensorID,
				SensorType: sensorState.Sensor.Type,
				Value:      sensorState.Sensor.Value,
			}
			if err := functions.WriteJSONLine(session.Conn, EnvelopeMessage); err != nil {
				log.Printf("Erro ao enviar sensor p/ cliente %s: %v", clientID, err)
			}
			session.writeMutex.Unlock()
		} else {
			log.Printf("Descartada telemetria do sensor %s para o cliente %s. Cliente ocupado.", sensorState.Sensor.SensorID, clientID)
		}
	}
}

// sendAllSensorsToNewClient envia o estado atual de todos os sensores para um novo cliente que acabou de se conectar.
// Params:
// session - A sessão do cliente para o qual os sensores serão enviados
// Returns:
// error   - Erro se algum envio apresentar falha
func sendAllSensorsToNewClient(session *ClientSession) error {
	sensorsMapMutex.RLock()
	sensorsCache := maps.Clone(sensorsMap) // Snapshot para evitar segurar o lock durante envios
	sensorsMapMutex.RUnlock()

	session.writeMutex.Lock() // Escrita atômica para evitar misturar mensagens
	defer session.writeMutex.Unlock()

	for _, sensorState := range sensorsCache {
		EnvelopeMessage := structs.SensorMessage{
			Type:       "sensor",
			SensorID:   sensorState.Sensor.SensorID,
			SensorType: sensorState.Sensor.Type,
			Value:      sensorState.Sensor.Value,
		}

		if err := functions.WriteJSONLine(session.Conn, EnvelopeMessage); err != nil {
			return err
		}
	}
	return nil
}

func pingClient(session *ClientSession) {

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		clientsMapMutex.RLock()
		currentSession, ok := clientsSessionMap[session.ID]
		if !ok || currentSession != session {
			clientsMapMutex.RUnlock()
			return
		}
		lastPong := session.LastPong
		clientsMapMutex.RUnlock()

		if time.Since(lastPong) > 4*time.Second {
			log.Printf("Cliente %s sem pong dentro do prazo; encerrando conexão", session.ID)
			_ = session.Conn.Close()
			return
		}

		session.writeMutex.Lock()
		err := functions.WriteJSONLine(session.Conn, structs.PingPongMessage{Type: "ping"})
		session.writeMutex.Unlock()
		if err != nil {
			log.Printf("Erro ao enviar ping para cliente %s: %v", session.ID, err)
			_ = session.Conn.Close()
			return
		}
	}
}
