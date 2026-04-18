package protocol

const (
	TypeConnectionAck = "connection_ack"
	TypeKeepAlive     = "ka"
	TypeUnsubscribe   = "unsubscribe"
	TypeSubscribe     = "subscribe"
	TypeData          = "data"
	TypePublish       = "publish"

	TypeSubscribeSuccess   = "subscribe_success"
	TypePublishSuccess     = "publish_success"
	TypeUnsubscribeSuccess = "unsubscribe_success"

	TypeError            = "error"
	TypeConnectionError  = "connection_error"
	TypeSubscribeError   = "subscribe_error"
	TypePublishError     = "publish_error"
	TypeUnsubscribeError = "unsubscribe_error"
)
