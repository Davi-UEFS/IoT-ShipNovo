package functions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

// Tempo máximo para tentativas de reconexão ao broker, com backoff exponencial.
const MAX_RETRY_TIME = 30 * time.Second
const MAX_RETRY_ATTEMPTS = 10

// ConnectWithRetry tenta se conectar ao endereço fornecido, com backoff exponencial em caso de falha.
// Params:
// addr - O endereço do broker para se conectar.
// Returns:
// net.Conn - A conexão estabelecida com o broker.
func ConnectWithRetry(addr string) net.Conn {
	dialerWithTimeout := net.Dialer{Timeout: 3 * time.Second}
	baseTime := time.Second

	for i := 0; i < MAX_RETRY_ATTEMPTS; i++ {
		conn, err := dialerWithTimeout.Dial("tcp", addr)
		if err == nil {
			return conn
		}

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			SafePrintf("Tempo de conexão esgotado, tentando novamente em %vs...\n", baseTime.Seconds())
		} else {
			SafePrintf("Broker indisponível (%v). Tentativa %d/%d. Tentando novamente em %vs...\n", err, i+1, MAX_RETRY_ATTEMPTS, baseTime.Seconds())
		}
		time.Sleep(baseTime)

		baseTime *= 2 // Backoff exponencial, com limite máximo
		if baseTime > MAX_RETRY_TIME {
			baseTime = MAX_RETRY_TIME
		}
	}

	SafeFatalf("Falha ao conectar ao broker após %d tentativas", MAX_RETRY_ATTEMPTS)
	return nil 
}

// WriteJSONLine escreve um valor como JSON e adiciona uma nova linha ao final, para leituras stream do TCP.
// Tem limite de tempo para escrita, para evitar bloqueios prolongados em conexões problemáticas.
// Params:
// conn - A conexão TCP onde os dados serão escritos.
// v - O valor a ser convertido para JSON e escrito na conexão.
// Returns:
// error - Um erro caso a operação de escrita falhe.
func WriteJSONLine(conn net.Conn, v any) error {
	buffer, err := json.Marshal(v)
	if err != nil {
		return err
	}
	buffer = append(buffer, '\n')

	// Define um prazo para a escrita e o limpa após a operação
	conn.SetWriteDeadline(time.Now().Add(time.Second))
	defer conn.SetWriteDeadline(time.Time{})

	_, err = conn.Write(buffer)
	return err
}

// DoHandshake faz o handshake com o broker numa conexão TCP, enviando uma mensagem de handshake.
// Aguarda um ACK (falho ou bem-sucedido) do broker para confirmar a conexão.
// Params:
// conn 	- A conexão TCP com o broker.
// reader 	- Um bufio.Reader para ler a resposta do broker.
// id 		- O ID (sensor ou atuador) que está se conectando.
// role 	- O papel (sensor ou atuador) do cliente que está se conectando.
// Returns:
// error 	- Um erro caso a operação de handshake falhe ou nil se for bem-sucedida.
func DoHandshake(conn net.Conn, reader *bufio.Reader, id string, role string) error {
	msg := structs.HandshakeMessage{
		Type: "handshake",
		Role: role,
		ID:   id}

	//WARNING: HANDSHAKE FICA COM 1S PQ DO WRITEJSONLINE
	if err := WriteJSONLine(conn, msg); err != nil {
		return err
	}

	var ack structs.HandshakeAck

	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // Tempo limite para receber o ACK do broker.
	defer conn.SetReadDeadline(time.Time{})

	line, err := reader.ReadBytes('\n')
	if err != nil {
		return err
	}
	if err := json.Unmarshal(line, &ack); err != nil {
		return err
	}

	if !ack.OK {
		if ack.Message == "" {
			return fmt.Errorf("Handshake recusado")
		}
		return fmt.Errorf("Handshake recusado: %s", ack.Message)
	}

	return nil
}

var printMu sync.Mutex // Mutex interno para escrita em stdout.

// SafePrintf é um Printf seguro para uso concorrente.
// Params:
// format 	- O formato da string de saída.
// a 		- Os argumentos para a string de saída.
func SafePrintf(format string, a ...any) {
	printMu.Lock()
	defer printMu.Unlock()
	fmt.Printf(format, a...)
}

// SafePrintln é um Println seguro para uso concorrente.
// Params:
// a 		- Os argumentos para a string de saída.
func SafePrintln(a ...any) {
	printMu.Lock()
	defer printMu.Unlock()
	fmt.Println(a...)
}

// SafeFatalf é um Fatalf seguro para uso concorrente.
// Params:
// format 	- O formato da string de saída.
// a 		- Os argumentos para a string de saída.
func SafeFatalf(format string, a ...any) {
	printMu.Lock()
	fmt.Printf(format, a...)
	printMu.Unlock()
	os.Exit(1)
}
