package main

import (
	"bufio"
	"encoding/json"
	"net"
	"time"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

// waitForBrokerMessages Rotina em paralelo que aguarda mensagens do broker na conexão.
// Params:
//
//	conn		- A conexão do cliente
//	reader 		- O leitor da conexão
//	brokerAddr 	- O endereço do broker, para reconexões
//	id		- O ID do cliente
func waitForBrokerMessages(conn net.Conn, reader *bufio.Reader, brokerAddr string, id string) {

	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		line, err := reader.ReadBytes('\n')
		if err != nil {

			functions.SafePrintf("Conexão com broker perdida: %v\n", err)

			time.Sleep(time.Second) //WARNING: RECONEXAO ACONTECE ANTES DO BROKER FECHAR. HARDCODED
			conn = functions.ConnectWithRetry(brokerAddr)
			if conn == nil {
				functions.SafeFatalf("Não foi possível estabelecer conexão com o broker em %s", brokerAddr)
			}

			reader = bufio.NewReader(conn)

			if err := functions.DoHandshake(conn, reader, id, "actuator"); err != nil {
				functions.SafeFatalf("Erro no handshake após reconexão: %v", err)
			}
			functions.SafePrintf("Handshake OK após reconexão (actuator id=%s)\n", id)
			sendActState(conn)
			continue

		}

		var env structs.EnvelopeMessage

		if err := json.Unmarshal(line, &env); err != nil {
			functions.SafePrintf("Mensagem com formato inválido do broker: %v\n", err)
			continue
		}

		if env.Type == "ping" {
			if err := functions.WriteJSONLine(conn, structs.PingPongMessage{Type: "pong"}); err != nil {
				functions.SafePrintf("Erro ao responder ping do broker: %v\n", err)
			}
			continue
		}

		var cmd structs.ActuatorCommand

		if err := json.Unmarshal(line, &cmd); err != nil {
			continue
		}

		handleCommand(cmd, conn)
	}
}
