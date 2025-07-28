package hub

import (
	go_json "github.com/goccy/go-json"
)

type WsMessage struct {
	Type    WsMessageType      `json:"type"`
	Payload go_json.RawMessage `json:"payload"`
}
