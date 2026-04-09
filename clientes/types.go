package main

import (
	"sync"
	"time"
)

// UIEvents representa um evento que deve ser mostrado na interface do usuário.
type UIEvents struct {
	Message string
}

// PendingAck representa o resultado de um ACK recebido do broker.
type PendingAck struct {
	OK      bool
	Message string
}

// Notifications mantém o estado das notificações para o usuário, incluindo o número de mensagens não lidas, a lista de mensagens e a última vez que as notificações foram atualizadas.
type Notifications struct {
	Mu         sync.Mutex
	Unread     int
	Messages   []string
	LastUpdate time.Time
}
