package eventsocket

import "encoding/json"

type MessageHandler func(data json.RawMessage)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}
