package main

import (
	"log"
	"net"
	"os"
)

// main inicia o broker, configurando as portas e iniciando os listeners TCP e UDP. Também inicia a rotina de monitoramento do TTL dos sensores.
// Params: Broker lê as portas TCP e UDP das variáveis de ambiente, ou usa valores padrão se não estiverem definidas.

func main() {
	udpPort := os.Getenv("UDP_PORT")
	tcpPort := os.Getenv("TCP_PORT")

	if udpPort == "" {
		udpPort = "9000"
	}
	if tcpPort == "" {
		tcpPort = "9001"
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ":"+udpPort)
	if err != nil {
		log.Fatalf("Endereço UDP inválido: %v", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Erro ao abrir UDP %v", err)
	}
	defer udpConn.Close()
	log.Printf("✅ Escutando UDP em :%s", udpPort)

	tcpLn, err := net.Listen("tcp", ":"+tcpPort)
	if err != nil {
		log.Fatalf("Erro ao abrir TCP %v", err)
	}
	defer tcpLn.Close()
	log.Printf("✅ Escutando TCP em :%s", tcpPort)

	go listenTCP(tcpLn)
	go listenUDPSensors(udpConn)
	go monitorSensorTTL()
	select {}					// Mantém o broker rodando indefinidamente
}
