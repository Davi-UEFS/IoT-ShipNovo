package main

import (
	"net"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

var act structs.Actuator

// handleCommand trata o comando recebido pelo broker, ligando ou desligando o atuador.
// Params: cmd - o comando do atuador
func handleCommand(cmd structs.ActuatorCommand, conn net.Conn) {

	switch cmd.Action {
	case "ligar":
		functions.SafePrintln("Ligando atuador")
		act.State = "ligado"
		act.LastAction = "Ligar"
	case "desligar":
		functions.SafePrintln("Desligando atuador")
		act.State = "desligado"
		act.LastAction = "Desligar"

	case "status":
		functions.SafePrintln("Enviando estado do atuador")
		// Aqui você pode implementar a lógica para enviar o estado atual do atuador para o broker
		// Por exemplo, você pode usar uma função para enviar o estado atual do atuador para o broker
		// sendActState(conn)
		sendActState(conn)
	default:
		functions.SafePrintf("Ação desconhecida: %q\n", cmd.Action)
	}

}

// sendActState envia o estado atual do atuador para o broker.
// Params: conn - a conexão com o broker
func sendActState(conn net.Conn) {

	if err := functions.WriteJSONLine(conn, act); err != nil {
		functions.SafePrintf("Erro ao enviar estado do atuador: %v\n", err)
	}
}

// initActuator inicia  um atuador com um ID e estado desligado.
// Params: id - o ID do atuador
// Returns: o atuador inicializado
func initActuator(id string) structs.Actuator {
	if id == "" {
		id = "exaustor-01"
	}

	return structs.Actuator{
		ID:         id,
		State:      "desligado",
		LastAction: "",
	}
}
