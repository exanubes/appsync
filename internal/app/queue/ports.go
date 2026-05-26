package queue

type ConnectionState interface {
	Done() <-chan struct{}
}
