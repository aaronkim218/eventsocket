package hubsocket

type BroadcastMessage[T any] struct {
	Client    *Client[T]
	WsMessage WsMessage
}
