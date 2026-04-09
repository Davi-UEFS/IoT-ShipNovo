package structs

// Sensor é o pacote JSON enviado ao broker via UDP.
// Todos os sensores usam a mesma estrutura — o broker diferencia pelo campo Tipo.
type Sensor struct {
	SensorID string  `json:"sensorID"`
	Type     string  `json:"sensorType"`
	Value    float64 `json:"value"`
}

// Actuator representa um atuador conectado ao broker.
// Um atuador pode ter vários estados, mas para simplicidade usamos apenas "ligado" e "desligado".
type Actuator struct {
	ID         string `json:"id"`
	State      string `json:"state"`
	LastAction string `json:"lastAction"`
}

// HandshakeAck é a resposta do broker ao handshake inicial de clientes e atuadores, indicando sucesso ou falha.
type HandshakeAck struct {
	Type    string `json:"type"` // "handshake_ack"
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

// HandshakeMessage é a mensagem inicial que clientes e atuadores enviam ao broker para se identificarem pelo TCP.
type HandshakeMessage struct {
	Type string `json:"type"` // "handshake"
	Role string `json:"role"`
	ID   string `json:"id"`
}

// EnvelopeMessage é a estrutura básica para identificar o tipo de mensagem recebida do broker.
type EnvelopeMessage struct {
	Type string `json:"type"`
}

// ClientCommand é a estrutura usada por clientes para enviar comandos aos atuadores via TCP.
type ClientCommand struct {
	Role       string `json:"role"`                 // "client"
	Type       string `json:"type"`                 // "command" | "query"
	ClientID   string `json:"id"`                   // client id
	RequestID  string `json:"request_id,omitempty"` // correlação com ACK
	ActuatorID string `json:"actuatorID"`
	Action     string `json:"action"`
}

// ActuatorCommand representa o comando que o broker envia para um atuador.
type ActuatorCommand struct {
	Action string `json:"action"`
}

// SensorMessage é a estrutura usada pelo broker/clientes para telemetria de sensores.
// O campo Type deve ser sempre "sensor" para diferenciar de outros tipos de mensagens.
type SensorMessage struct {
	Type       string  `json:"type"` // "sensor"
	SensorID   string  `json:"sensorID"`
	SensorType string  `json:"sensorType"`
	Value      float64 `json:"value"`
}

// ActuatorMessage é a estrutura usada pelo broker/clientes para telemetria de atuadores.
// O campo Type deve ser sempre "actuator" para diferenciar de outros tipos de mensagens.
type ActuatorMessage struct {
	Type       string `json:"type"` // "actuator"
	ID         string `json:"id"`
	State      string `json:"state"`
	LastAction string `json:"lastAction"`
}

// ACKMessage é enviado pelo broker ao cliente para confirmar (ou rejeitar) um comando.
type ACKMessage struct {
	Type       string `json:"type"`                 // "ack"
	RequestID  string `json:"request_id"`           // correlação com o request original
	OK         bool   `json:"ok"`                   // true = sucesso, false = falha
	Message    string `json:"message,omitempty"`    // motivo em caso de falha
	ActuatorID string `json:"actuatorID,omitempty"` // atuador alvo do comando
	Action     string `json:"action,omitempty"`     // ação executada
}

// RemoveMessage é a estrutura usada pelo broker para notificar clientes sobre a remoção de sensores ou atuadores.
type RemoveMessage struct {
	Type   string `json:"type"`   // "remove"
	Entity string `json:"entity"` // "sensor" | "actuator"
	ID     string `json:"id"`
	Reason string `json:"reason,omitempty"` // "disconnect" | "expired" | ...
}

type PingPongMessage struct {
	Type string `json:"type"` // "ping" | "pong"
}
