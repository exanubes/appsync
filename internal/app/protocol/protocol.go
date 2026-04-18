package protocol

type KeepAliveMessage struct{}

type ConnectionAckMessage struct {
	TimeoutMs int
}

type ErrorMessage struct {
	ID     string
	Errors []ErrorMetadata
}

type ErrorMetadata struct {
	Type    string
	Message string
}
