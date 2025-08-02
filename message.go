package eventsocket

import (
	go_json "github.com/goccy/go-json"
)

type MessageHandler func(data go_json.RawMessage)

type Message struct {
	Type string             `json:"type"`
	Data go_json.RawMessage `json:"data"`
}
