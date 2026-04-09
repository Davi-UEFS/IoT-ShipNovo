package main

import (
	"net"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

// sendCommandToActuator envia ao broker um comando para um atuador.
// Params:
// conn 		- a conexão com o broker
// clientID 	- o ID do cliente que está enviando o comando
// requestID 	- o ID da requisição, para controle de ACKs pendentes
// actuatorID 	- o ID do atuador que deve receber o comando
// action 		- a ação a ser executada pelo atuador (ex: "ligar", "desligar")
// Returns:
// error 		- erro caso ocorra algum problema ao enviar o comando
func sendCommandToActuator(conn net.Conn, clientID, requestID, actuatorID, action string) error {
	cmd := structs.ClientCommand{
		Role:       "client",
		Type:       "command",
		ClientID:   clientID,
		RequestID:  requestID,
		ActuatorID: actuatorID,
		Action:     action,
	}
	return functions.WriteJSONLine(conn, cmd)
}

// askBrokerForActuators pede ao broker a lista de atuadores registrados.
// Params:
// conn 		- a conexão com o broker
// clientID 	- o ID do cliente que está solicitando a lista de atuadores
// requestID 	- o ID da requisição, para controle de ACKs pendentes
// Returns:
// error 		- erro caso ocorra algum problema ao enviar a solicitação
func askBrokerForActuators(conn net.Conn, clientID string, requestID string) error {
	cmd := structs.ClientCommand{
		Role:       "client",
		Type:       "query",
		ClientID:   clientID,
		RequestID:  requestID,
		ActuatorID: "-",
		Action:     "-",
	}
	return functions.WriteJSONLine(conn, cmd)
}
