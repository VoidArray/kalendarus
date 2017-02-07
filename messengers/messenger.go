package messengers

type Messenger interface {
	Send(message string) error
}
