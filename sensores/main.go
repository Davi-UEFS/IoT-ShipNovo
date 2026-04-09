package main

import (
	"log"
	"net"
	"os"
)

// ── Main ──────────────────────────────────────────────────────────────────────

// main inicializa o sensor e estabelece a conexão UDP com o broker.
// Params: Usa variáveis de ambiente para configurar o ID do sensor, tipo e endereço do broker. Valores padrão são usados se as variáveis não estiverem definidas.
func main() {
	id := os.Getenv("SENSOR_ID")
	tipo := os.Getenv("SENSOR_TIPO")
	brokerAddr := os.Getenv("BROKER_IP")

	// Valores padrão para rodar fora do Docker
	if id == "" {
		id = "sensor-teste-01"
	}
	if tipo == "" {
		tipo = "temperatura"
	}
	if brokerAddr == "" {
		brokerAddr = "localhost:9000"
	}

	if !checkSensorType(tipo) {
		log.Fatalf("Tipo de sensor desconhecido: %q", tipo)   // Fatal se o tipo dado não for reconhecido
	}

	addr, err := net.ResolveUDPAddr("udp", brokerAddr)
	if err != nil {
		log.Fatalf("Endereço inválido %s: %v", brokerAddr, err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("Falha ao abrir UDP: %v", err)
	}
	defer conn.Close()

	publishSensorLoop(conn, id, tipo)
}
