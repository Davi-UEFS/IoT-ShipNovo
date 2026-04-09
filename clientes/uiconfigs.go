package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Davi-UEFS/IoT-Ship/shared/functions"
	"github.com/Davi-UEFS/IoT-Ship/shared/structs"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// ShowSensorsTable exibe uma tabela formatada com os dados dos sensores.
// Params:
// sensors 	- A lista de sensores a ser exibida
func ShowSensorsTable(sensors []structs.Sensor) {
	tb := table.NewWriter()
	tb.Style().Box = (table.StyleBoxBold) // Recomenda-se ler doc da biblioteca para entender as opções de estilo.
	tb.Style().Options.SeparateColumns = true
	tb.Style().Options.SeparateHeader = true
	tb.Style().Options.SeparateFooter = true
	tb.Style().Format.Header = text.FormatUpper
	tb.Style().Format.Footer = text.FormatDefault
	tb.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignCenter},
		{Number: 2, Align: text.AlignCenter},
		{Number: 3, Align: text.AlignCenter},
	})

	tb.AppendHeader(table.Row{"Sensor ID", "Tipo", "Valor"})

	rows := sensorsToTableRows(sensors)
	tb.AppendRows(rows)
	tb.AppendFooter(table.Row{"Pressione ENTER para voltar"})
	renderCenteredTable(tb.Render(), 100)
}

// ShowActuatorsTable exibe uma tabela formatada com os dados dos atuadores.
// Params:
// actuators 	- A lista de atuadores a ser exibida
func ShowActuatorsTable(actuators []structs.Actuator) {
	tb := table.NewWriter()
	tb.Style().Box = (table.StyleBoxBold)
	tb.Style().Options.SeparateColumns = true
	tb.Style().Options.SeparateHeader = true
	tb.Style().Options.SeparateFooter = true
	tb.Style().Format.Header = text.FormatUpper
	tb.Style().Format.Footer = text.FormatDefault
	tb.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignCenter},
		{Number: 2, Align: text.AlignCenter},
		{Number: 3, Align: text.AlignCenter},
	})

	tb.AppendHeader(table.Row{"Atuador ID", "Estado", "Última ação"})

	tb.SortBy([]table.SortBy{ // Ordena por ID (crescente)
		{Name: "Atuador ID", Mode: table.Asc, IgnoreCase: true},
	})
	rows := actuatorsToTableRows(actuators)
	tb.AppendRows(rows)
	tb.AppendFooter(table.Row{"Pressione ENTER para voltar"})
	renderCenteredTable(tb.Render(), 100)
}

// sensorsToTableRows converte uma lista de sensores para o formato de linhas da tabela.
// Params:
// sensors 	- A lista de sensores a ser convertida
// Returns:
// []table.Row - A lista de linhas formatadas para a tabela
func sensorsToTableRows(sensors []structs.Sensor) []table.Row {
	sensorsCopy := append([]structs.Sensor(nil), sensors...)
	sort.Slice(sensorsCopy, func(i, j int) bool {
		typeI := strings.ToLower(strings.TrimSpace(sensorsCopy[i].Type))
		typeJ := strings.ToLower(strings.TrimSpace(sensorsCopy[j].Type))
		if typeI == typeJ {
			return strings.ToLower(sensorsCopy[i].SensorID) < strings.ToLower(sensorsCopy[j].SensorID)
		}
		return typeI < typeJ
	})

	rows := make([]table.Row, len(sensorsCopy))

	for rowIndex, sensor := range sensorsCopy {
		rows[rowIndex] = table.Row{sensor.SensorID, formatSensorType(sensor.Type), fmt.Sprintf("%.2f", sensor.Value)}
	}
	return rows
}

// actuatorsToTableRows converte uma lista de atuadores para o formato de linhas da tabela.
// Params:
// actuators 	- A lista de atuadores a ser convertida
// Returns:
// []table.Row - A lista de linhas formatadas para a tabela
func actuatorsToTableRows(actuators []structs.Actuator) []table.Row {

	rows := make([]table.Row, len(actuators))

	for rowIndex, actuator := range actuators {
		rows[rowIndex] = table.Row{actuator.ID, formatActuatorState(actuator.State), actuator.LastAction}
	}
	return rows
}

func formatActuatorState(state string) string {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "ligado":
		return "\x1b[1;32mLigado\033[0m"
	case "desligado":
		return "\x1b[1;31mDesligado\033[0m"
	default:
		return state
	}
}

func formatSensorType(sensorType string) string {
	switch strings.ToLower(strings.TrimSpace(sensorType)) {
	case "temperatura":
		return "\033[38;5;208mTemperatura\033[0m"
	case "combustivel":
		return "\033[38;5;178mCombustível\033[0m"
	case "porcao":
		return "\033[38;5;117mPorão\033[0m"
	default:
		return sensorType
	}
}

func renderCenteredTable(rendered string, targetWidth int) {
	if targetWidth <= 0 {
		targetWidth = 100
	}

	lines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
	maxWidth := 0
	for _, line := range lines {
		width := text.StringWidth(line)
		if width > maxWidth {
			maxWidth = width
		}
	}

	leftPadding := 0
	if targetWidth > maxWidth {
		leftPadding = (targetWidth - maxWidth) / 2
	}
	padding := strings.Repeat(" ", leftPadding)

	for _, line := range lines {
		functions.SafePrintln(padding + line)
	}
}
