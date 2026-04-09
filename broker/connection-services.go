package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

// listenTCP aceita conexões TCP e inicia uma goroutine para cada conexão para lidar com o handshake e registro.
// Params:
// ln - O listener TCP para aceitar conexões
func listenTCP(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("erro ao aceitar conexão %v", err)
			continue
		}
		go handleConn(conn)
	}
}

// handleConn trata a conexão aceita e encaminha para o registro apropriado.
// Params:
// conn - A conexão TCP aceita do cliente ou atuador
func handleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Ignora conexões que não enviarem o handshake em 5 segundos
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	line, err := reader.ReadBytes('\n')
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Printf("Handshake não recebido em 5s de %s", conn.RemoteAddr().String())
		}
		return
	}

	conn.SetReadDeadline(time.Time{})

	var hello structs.HandshakeMessage
	if err := json.Unmarshal(line, &hello); err != nil {
		log.Printf("Handshake inválido de %s", conn.RemoteAddr().String())
		return
	}
	if hello.ID == "" || (hello.Role != "client" && hello.Role != "actuator") {
		log.Printf("Handshake incompleto de %s: %s", conn.RemoteAddr().String(), string(line))
		return
	}

	duplicatedID := checkDuplicateID(hello.ID)
	var ack structs.HandshakeAck

	if duplicatedID {
		ack = structs.HandshakeAck{
			Type:    "handshake_ack",
			OK:      false,
			Message: "ID já em uso, escolha outro ID para conectar: " + hello.ID,
		}

	} else {
		ack = structs.HandshakeAck{
			Type: "handshake_ack",
			OK:   true,
		}
	}
	if err := functions.WriteJSONLine(conn, ack); err != nil {
		log.Printf("Erro ao enviar handshake ack para %s: %v", conn.RemoteAddr().String(), err)
		return
	}

	if !ack.OK {
		log.Printf("Conexão rejeitada (%s id=%s): %s", hello.Role, hello.ID, ack.Message)
		return
	}

	//TODO: FECHAR READER E ABRIR NO REGISTER
	switch hello.Role {
	case "actuator":
		pendingActuatorsUpdate.Add(1) // Faz atualização pois espera dados após o handshake
		registerActuator(conn, reader, hello.ID)
	case "client":
		registerClient(conn, reader, hello.ID)
	}
}

// checkDuplicateID verifica se o ID fornecido já está em uso por outro cliente ou atuador conectado.
// Params: id - O ID a ser verificado
// Returns:
// true se o ID estiver em uso, false caso contrário
func checkDuplicateID(id string) bool {
	actuatorsMapMutex.RLock()
	_, inActuators := actuatorsMap[id]
	actuatorsMapMutex.RUnlock()

	clientsMapMutex.RLock()
	_, inClients := clientsSessionMap[id]
	clientsMapMutex.RUnlock()

	return inActuators || inClients
}


