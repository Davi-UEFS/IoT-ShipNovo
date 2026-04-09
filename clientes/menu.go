package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
)

// clearScreen limpa o terminal dependendo do SO.
func clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}

// clientMenu é o loop principal do cliente, exibindo o menu e tratando as opções escolhidas pelo usuário.
// Params:
// clientID - O ID do cliente, usado para identificar as requisições e comandos enviados ao broker.
func clientMenu(clientID string) {

	for {

		if getOfflineStatus() {
			continue
		}
		renderMainMenu() // Exibe opções do menu principal

		choice, interrupted, err := readMenuChoice()
		if interrupted { // Se a leitura foi interrompida por uma notificação, exibe o menu atualizado.
			continue
		}
		if err != nil {
			functions.SafePrintln("\nEntrada inválida. Digite um número de 1 a 6.")
			pauseEnter()
			clearScreen()
			continue
		}

		switch choice {
		case 1: // Pedir atuadores
			requestID := newCommandID(clientID)    // Gera ID de requisição
			ackCh := registerPendingAck(requestID) // Registra canal para receber ACK do broker
			conn := getBrokerConn()
			if conn == nil {
				functions.SafePrintln("Sem conexão com o broker no momento. Aguarde a reconexão automática.")
				unregisterPendingAck(requestID)
				pauseEnter()
				clearScreen()
				continue
			}

			if err := askBrokerForActuators(conn, clientID, requestID); err != nil {
				functions.SafePrintf("Erro ao solicitar atuadores: %v\n", err)
			} else {
				select {
				case ack := <-ackCh:
					if ack.OK {
						functions.SafePrintln("\nSolicitação de atuadores aceita pelo broker")
					} else {
						functions.SafePrintf("Erro no ACK: %s\n", ack.Message)
					}
				case <-time.After(600 * time.Millisecond):
					functions.SafePrintf("ACK do query (id=%s) demorou muito para responder\n", requestID)
					functions.SafePrintln("Recomenda-se verificar estado dos atuadores antes de tentar novamente")
				}
			}

			unregisterPendingAck(requestID) // Limpa o canal de ACK pendente
			pauseEnter()
			clearScreen()

		case 2: // Mostrar atuadores
			actuatorsMapMutex.RLock()
			// Cópia os atuadores para uma slice para exibição, evitando segurar o RLock durante a renderização
			actuators := make([]structs.Actuator, 0, len(actuatorsMap))
			for _, act := range actuatorsMap {
				actuators = append(actuators, act)
			}
			actuatorsMapMutex.RUnlock()
			ShowActuatorsTable(actuators)
			pauseEnter()
			clearScreen()

		case 3: // Enviar comando
			sendCommandMiniMenu(clientID)

		case 4: // Mostrar sensores
			liveSensorsView()

		case 5:
			showNotifications()

		case 6: // Sair
			functions.SafePrintln("Saindo...")
			return

		default:
			functions.SafePrintln("Opção inválida. Escolha entre 1 e 6.")
			pauseEnter()
		}
	}
}

// renderMainMenu mostra as opções do menu principal.
// Se houver notificações não lidas, exibe um ícone de sino amarelo e a quantidade de notificações pendentes.
func renderMainMenu() {
	unread, _, _ := getNotificationsSnapshot()

	clearScreen()
	functions.SafePrintln("Selecione uma opção")
	functions.SafePrintln("1) Pedir atuadores")
	functions.SafePrintln("2) Mostrar atuadores")
	functions.SafePrintln("3) Enviar comando para atuador")
	functions.SafePrintln("4) Mostrar sensores")
	if unread > 0 {
		fmt.Printf("5) \033[33m🔔 Ver notificações (%d)\033[0m\n", unread)
	} else {
		functions.SafePrintln("5) Ver notificações")
	}
	functions.SafePrintln("6) Sair")
}

// sendCommandMiniMenu é o minimenu para enviar um comando a um atuador.
// Params:
// clientID - O ID do cliente, usado para identificar a requisição e o comando enviado ao broker.
func sendCommandMiniMenu(clientID string) {
	empty := false

	actuatorsMapMutex.RLock()
	if len(actuatorsMap) == 0 {
		empty = true
	}
	actuatorsMapMutex.RUnlock()

	if empty {
		functions.SafePrintln("Nenhum atuador disponível para enviar comando.")
		functions.SafePrintln("Solicite atuadores na opção 1 e verifique o resultado nas notificações")
		pauseEnter()
		clearScreen()
		return
	}

	actuatorID, err := readLine("ID do atuador: ")
	if err != nil {
		functions.SafePrintf("Erro ao ler ID do atuador: %v\n", err)
		return
	}
	if actuatorID == "" {
		functions.SafePrintln("ID do atuador não pode ser vazio.")
		pauseEnter()
		clearScreen()
		return
	}

	clearScreen()
	functions.SafePrintln("Escolha a ação")
	functions.SafePrintln("1) Ligar")
	functions.SafePrintln("2) Desligar")

	actionChoice, err := readIntOption("Ação: ")
	if err != nil || (actionChoice != 1 && actionChoice != 2) {
		functions.SafePrintln("Ação inválida.")
		pauseEnter()
		clearScreen()
		return
	}

	action := "desligar" // Converte a escolha da ação para o formato esperado pelo broker ("ligar" ou "desligar")
	if actionChoice == 1 {
		action = "ligar"
	}

	requestID := newCommandID(clientID)    // Gera ID de requisição para controle de ACKs pendentes
	ackCh := registerPendingAck(requestID) // Registra canal para receber ACK do broker
	conn := getBrokerConn()
	if conn == nil {
		functions.SafePrintln("Sem conexão com o broker no momento. Aguarde a reconexão automática.")
		unregisterPendingAck(requestID)
		pauseEnter()
		clearScreen()
		return
	}

	if err := sendCommandToActuator(conn, clientID, requestID, actuatorID, action); err != nil {
		functions.SafePrintf("Erro ao encaminhar comando para o broker %s: %v\n", actuatorID, err)
	} else {
		select {
		case ack := <-ackCh:
			if ack.OK {
				functions.SafePrintf("Comando aceito: \nAtuador = %s \nAção = %s\n", actuatorID, action)
			} else {
				functions.SafePrintf("Comando recusado: \nAtuador = %s \nMotivo = %s\n", actuatorID, ack.Message)
			}
		case <-time.After(600 * time.Millisecond):
			functions.SafePrintf("ACK do comando (id=%s) demorou muito para responder\n", requestID)
		}
	}

	unregisterPendingAck(requestID) // Limpa o canal de ACK pendente

	pauseEnter()
	clearScreen()
}

// liveSensorsView exibe os sensores em tempo real, atualizando a cada 300ms.
func liveSensorsView() {
	stop := make(chan struct{})

	go func() {
		// ENTER para sair da visualização
		_, _ = readLine("")
		close(stop)
	}()

	ticker := time.NewTicker(300 * time.Millisecond) // Atualiza a visualização a cada 300ms
	defer ticker.Stop()

	for {
		select {
		case <-stop: // User aperta ENTER
			return
		case <-ticker.C: // Atualiza a visualização
			clearScreen()

			sensorsMapMutex.RLock()
			sensors := make([]structs.Sensor, 0, len(sensorsMap)) // Snapshot
			for _, s := range sensorsMap {
				sensors = append(sensors, s)
			}
			sensorsMapMutex.RUnlock()

			ShowSensorsTable(sensors)
			functions.SafePrintln("\nPressione ENTER para voltar...")
		}
	}
}
