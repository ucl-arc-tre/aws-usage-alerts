package sns

type Interface interface {
	Send(content string) error
}
