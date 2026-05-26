package pending

type ConnectionState interface {
	Done() <-chan struct{}
}
