package eventsocket

import "encoding/json"

type CreateClientHandler func(client *Client)

type RemoveClientHandler func(clientID string)

type JoinRoomHandler func(roomID string, clientID string)

type LeaveRoomHandler func(roomID string, clientID string)

type NewRoomHandler func(roomID string)

type DeleteRoomHandler func(roomID string)

type MessageHandler func(data json.RawMessage)
