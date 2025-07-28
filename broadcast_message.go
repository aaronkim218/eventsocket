package hubsocket

type BroadcastMessage[T any] struct {
	client    *Client[T]
	wsMessage WsMessage
}
