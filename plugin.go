package hubsocket

type ClientJoinPlugin[T any] interface {
	HandleClientJoin(*ActiveRoom[T], *Client[T]) error
}

type BroadcastMessagePlugin[T any] interface {
	MessageType() WsMessageType
	HandleBroadcastMessage(*ActiveRoom[T], BroadcastMessage[T]) error
}

type ClientLeavePlugin[T any] interface {
	HandleClientLeave(*ActiveRoom[T], *Client[T]) error
}
