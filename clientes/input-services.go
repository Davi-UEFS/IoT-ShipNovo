package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
)

// startStdinReader cria uma goroutine para ler o teclado.
// Envia a entrada para o canal stdinLines, que é lido pelas funções de input.
// Params: EXPLICACAO AQUI.
// Returns: EXPLICACAO AQUI.
func startStdinReader() {
	stdinReaderOnce.Do(func() {
		go func() {
			reader := bufio.NewReader(os.Stdin)
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					close(stdinLines)
					return
				}
				stdinLines <- line
			}
		}()
	})
}

// readLine exibe um prompt e lê uma linha do teclado, retornando a string trimmed.
// Params:
// prompt 	- A mensagem a ser exibida antes da leitura

// Returns:
// string 	- A entrada do usuário, sem espaços extras.
// error 	- Erro se a leitura falhar.
func readLine(prompt string) (string, error) {
	startStdinReader()
	if prompt != "" {
		fmt.Print(prompt)
	}
	line, ok := <-stdinLines
	if !ok {
		return "", io.EOF
	}
	return strings.TrimSpace(line), nil
}

// readIntOption lê uma linha e converte para inteiro.
// Params: 
// prompt 	- A mensagem a ser exibida antes da leitura
// Returns: 
// int 		- O valor inteiro lido
// error 	- Erro se a leitura ou conversão falharem
func readIntOption(prompt string) (int, error) {
	line, err := readLine(prompt)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(line)
}

// readMenuChoice lê a escolha do menu.
// Pode ser interrompida se uma notificação de atualização chegar.
// Returns: 
// choice 		- A escolha do menu
// interrupted 	- true se a leitura foi interrompida por uma notificação
// error 		- Erro se a leitura falhar
func readMenuChoice() (choice int, interrupted bool, err error) {
	startStdinReader()			// A goroutine para ler a entrada.
	fmt.Print("Escolha: ")

	select {
	case <-notificationsUpdateCh:
		return 0, true, nil
	case line, ok := <-stdinLines:
		if !ok {
			return 0, false, os.ErrClosed
		}
		line = strings.TrimSpace(line)
		choice, err := strconv.Atoi(line)
		if err != nil {
			return 0, false, err
		}
		return choice, false, nil
	}
}

// pauseEnter é usado para pausar a execução até que o usuário pressione ENTER.
func pauseEnter() {
	functions.SafePrintln("\nPressione ENTER para continuar...")
	_, _ = readLine("")
}
