package main

import "fmt"

// newCommandID gera um ID de comando único combinando o clientID com um contador atômico.
// Params:
// clientID - o ID do cliente para o qual o comando está sendo gerado.
// Returns:
// Uma string única representando o ID do comando, no formato "clientID-cmd-n"
func newCommandID(clientID string) string {
	n := commandCounter.Add(1)
	return fmt.Sprintf("%s-cmd-%d", clientID, n)
}

// registerPendingAck cria e armazena no map um canal de ACK para o dado requestID.
// Params:
// requestID - o ID da requisição para a qual o ACK está sendo registrado.
// Returns: O canal de ACK criado.
func registerPendingAck(requestID string) chan PendingAck {
	ch := make(chan PendingAck, 1)
	pendingAcksMutex.Lock()
	pendingAcks[requestID] = ch
	pendingAcksMutex.Unlock()
	return ch
}

// unregisterPendingAck remove do map e fecha o canal de ACK associado ao requestID.
// Params:
// requestID - o ID da requisição a ser removido do map.
func unregisterPendingAck(requestID string) {
	pendingAcksMutex.Lock()
	if ch, ok := pendingAcks[requestID]; ok {
		delete(pendingAcks, requestID)
		close(ch)
	}
	pendingAcksMutex.Unlock()
}

// deliverAck entrega um ACK ao canal pendente associado ao requestID (se existir).
// Params:
// requestID 	- o ID da requisição para a qual o ACK deve ser entregue.
// ack 			- o ACK a ser entregue, contendo informações sobre o resultado do comando.
func deliverAck(requestID string, ack PendingAck) {
	pendingAcksMutex.Lock()
	ch, ok := pendingAcks[requestID]
	pendingAcksMutex.Unlock()

	if ok {
		ch <- ack
	}
}
