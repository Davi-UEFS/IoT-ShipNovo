package main

import (
	"net"
	"sync"
	"time"

	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

// Não importa Actuator do shared porque a estrutura do broker é diferente da do cliente (campo Conn)
type Actuator struct {
	ID         string     `json:"id"`
	State      string     `json:"state"`
	LastAction string     `json:"lastAction"`
	Conn       net.Conn   `json:"-"`
	writeMutex sync.Mutex `json:"-"`
	LastPong   time.Time  `json:"-"`
}

// SensorState mantém o estado do sensor e a última vez que ele foi visto, para monitoramento de TTL.
type SensorState struct {
	Sensor   structs.Sensor
	LastSeen time.Time
}

// ClientSession mantém a sessão do cliente, incluindo o ID, a conexão e um mutex para escrita segura.
type ClientSession struct {
	ID         string
	Conn       net.Conn
	writeMutex sync.Mutex
	LastPong   time.Time
}

// Tempo de vida dos sensores
const TTLSENSOR = time.Second * 20

var (
	sensorsMap             = map[string]SensorState{}
	actuatorsMap           = map[string]*Actuator{}
	clientsSessionMap      = map[string]*ClientSession{}
	sensorsMapMutex        sync.RWMutex
	actuatorsMapMutex      sync.RWMutex
	clientsMapMutex        sync.RWMutex
	pendingActuatorsUpdate sync.WaitGroup
)
