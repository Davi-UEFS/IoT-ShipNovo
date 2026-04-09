package main

import (
	"bufio"
	"log"
	"os"
	"time"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
)

// main inicia a conexão com o broker, faz o handshake e inicia as goroutines para receber mensagens/notificações e exibe o menu.
// Params: Usa as variáveis de ambiente BROKER_ADDR e CLIENT_ID para configuração. Se não estiverem definidas, usa valores padrão.
func main() {
	brokerAddr := os.Getenv("BROKER_ADDR")
	clientID := os.Getenv("CLIENT_ID")

	if brokerAddr == "" {
		brokerAddr = "localhost:9001"
	}

	if clientID == "" {
		clientID = "cliente-01"
	}

	conn := functions.ConnectWithRetry(brokerAddr)
	if conn == nil {
		log.Fatalf("Não foi possível estabelecer conexão com o broker em %s", brokerAddr)
	}
	setBrokerConn(conn)
	reader := bufio.NewReader(conn)
	defer conn.Close()
	log.Printf("Conectado ao broker: %s", brokerAddr)

	if err := functions.DoHandshake(conn, reader, clientID, "client"); err != nil {
		log.Fatalf("Erro no handshake do cliente: %v", err)
	}
	log.Printf("Handshake OK (client id=%s)", clientID)

	functions.SafePrintln("Abrindo menu...")
	time.Sleep(2 * time.Second)

	go startEventListener()
	go receiveBrokerMessages(reader, brokerAddr, clientID)

	clientMenu(clientID)
}
