package main

import (
	"os/exec"
	"slices"
	"time"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
)

// startEventListener é uma goroutine que escuta o canal de eventos.
// Consome os eventos do canal e guarda na struct global de notificações.
// Dispara um sinal de atualização de notificações sempre que um novo evento chegar.
func startEventListener() {
	for event := range eventsChannel {
		notifications.Mu.Lock()
		notifications.Unread++
		notifications.Messages = append(notifications.Messages, event.Message)
		notifications.LastUpdate = time.Now()
		notifications.Mu.Unlock()

		select {
		case notificationsUpdateCh <- struct{}{}:
			if jingleCooldown.TryLock() {
				go playJingle() // Ignora tentar tocar o jingle se ele já estiver tocando.
			}
		default:
		}
	}
}

// showNotifications mostra as notificações para o usuário e limpa as notificações após a visualização.
func showNotifications() {
	unread, messages, updatedAt := getNotificationsSnapshot()

	if unread == 0 {
		functions.SafePrintf("\nVocê não tem notificações. Atualizado em %s\n", updatedAt.Format("15:04:05"))
		pauseEnter()
		clearScreen()
		return
	}

	functions.SafePrintf("\nVocê tem \033[33m %d \033[0m novas mensagens! Atualizado em %s\n", unread, updatedAt.Format("15:04:05"))
	for _, msg := range messages {
		functions.SafePrintf("- %s\n", msg)
	}

	clearNotifications()
	pauseEnter()
	clearScreen()
}

// getNotificationsSnapshot retorna uma cópia dos dados de notificações atuais.
// Returns:
// unread 		- O número de notificações não lidas
// messages 	- Uma cópia da lista de mensagens de notificações
// updatedAt 	- A data/hora da última atualização das notificações
func getNotificationsSnapshot() (unread int, messages []string, updatedAt time.Time) {
	notifications.Mu.Lock()
	defer notifications.Mu.Unlock()

	messagesCopy := slices.Clone(notifications.Messages)
	return notifications.Unread, messagesCopy, notifications.LastUpdate
}

// clearNotifications reseta a struct de notificações.
func clearNotifications() {
	notifications.Mu.Lock()
	defer notifications.Mu.Unlock()

	notifications.Unread = 0
	notifications.Messages = nil
}

func playJingle() error {
	defer jingleCooldown.Unlock()
	cmd := exec.Command(
		"aplay",
		"-q",
		"-D", "plughw:CARD=Generic_1,DEV=0",
		"/app/clientes/jingle.wav",
	)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}
