package protocol

type ConnectionAckMessage struct {
	Type      string `json:"type"`
	TimeoutMs int    `json:"connectionTimeoutMs"`
}

type ErrorMessage struct {
	Type   string          `json:"type"`
	ID     string          `json:"id"`
	Errors []ErrorMetadata `json:"errors"`
}

type ErrorMetadata struct {
	Type    string `json:"errorType"`
	Message string `json:"message"`
}
