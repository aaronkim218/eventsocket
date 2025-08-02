package eventsocket

import "errors"

var (
	ErrClientNotFound = errors.New("eventsocket: client not found")
	ErrRoomNotFound   = errors.New("eventsocket: room not found")
	ErrClientExists   = errors.New("eventsocket: client already exists with provided ID")
)
