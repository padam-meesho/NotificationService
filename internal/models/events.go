package models

import "encoding/json"

type SendSmsPayload struct {
	MessageId string `json:"message_id"` // this shall be a unique uuid
}

type KafkaPayload struct {
	Type string          `json:"type"` // this tells us which type of payload is being consumed.
	Data json.RawMessage `json:"data"` // this shall be further consumed
}
