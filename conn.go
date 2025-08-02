package eventsocket

type Conn interface {
	ReadJSON(any) error
	WriteJSON(v any) error
	Close() error
}
