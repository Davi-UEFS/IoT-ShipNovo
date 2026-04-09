package main

import (
	"bufio"
	"log"
	"os"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
)

// main cria a conexão e reader global e inicia o atuador.
// Tenta fazer handshake e fecha processo se falhar.

func main() {
	id := os.Getenv("ACTUATOR_ID")
	brokerAddr := os.Getenv("BROKER_ADDR")

	if id == "" {
		id = "exaustor-01"
	}
	if brokerAddr == "" {
		brokerAddr = "localhost:9001"
	}

	act = initActuator(id)

	conn := functions.ConnectWithRetry(brokerAddr)
	if conn == nil {
		log.Fatalf("Não foi possível estabelecer conexão com o broker em %s", brokerAddr)
	}

	defer conn.Close()
	reader := bufio.NewReader(conn)

	if err := functions.DoHandshake(conn, reader, id, "actuator"); err != nil {
		log.Fatalf("Erro no handshake: %v", err)
	} else {
		log.Printf("Handshake OK (actuator id=%s)", id)
		sendActState(conn)
	}

	waitForBrokerMessages(conn, reader, brokerAddr, id)
}
