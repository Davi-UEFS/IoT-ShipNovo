package main

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

var (
	sensorsMap        = make(map[string]structs.Sensor)
	actuatorsMap      = make(map[string]structs.Actuator)
	sensorsMapMutex   sync.RWMutex
	actuatorsMapMutex sync.RWMutex
	eventsChannel     = make(chan UIEvents, 30) // Canal para eventos de UI.

	// pendingAcks mapeia request_id para canal de ACK pendente.
	pendingAcks      = make(map[string]chan PendingAck)
	pendingAcksMutex sync.Mutex

	// commandCounter é usado para gerar IDs únicos de comando.
	commandCounter atomic.Uint64

	brokerConnMutex sync.RWMutex
	brokerConn      net.Conn

	notifications         Notifications            // Struct global para armazenar notificações e seu estado.
	notificationsUpdateCh = make(chan struct{}, 1) // Canal para sinalizar atualização de notificações.
	stdinLines            = make(chan string, 1)   // Canal para leitura do teclado
	stdinReaderOnce       sync.Once                // Garante que a goroutine de leitura do teclado seja iniciada apenas uma vez.
	jingleCooldown        sync.Mutex
	isOffline             bool
	isOfflineMutex        sync.RWMutex
)

// getBrokerConn retorna a conexão TCP atual do cliente com o broker.
func getBrokerConn() net.Conn {
	brokerConnMutex.RLock()
	defer brokerConnMutex.RUnlock()
	return brokerConn
}

// setBrokerConn troca a conexão atual usada pelo cliente.
// Se existir uma conexão antiga diferente da nova, ela é fechada.
func setBrokerConn(newConn net.Conn) {
	brokerConnMutex.Lock()
	oldConn := brokerConn
	brokerConn = newConn
	brokerConnMutex.Unlock()

	if oldConn != nil && oldConn != newConn {
		_ = oldConn.Close()
	}
}

func getOfflineStatus() bool {
	isOfflineMutex.RLock()
	defer isOfflineMutex.RUnlock()

	return isOffline
}

func setOfflineStatus(status bool) {
	isOfflineMutex.Lock()
	isOffline = status
	isOfflineMutex.Unlock()
}