package subscription

type Subscription struct {
	id      string
	channel string
}

func New(sub_id string, channel string) *Subscription {
	return &Subscription{
		id:      sub_id,
		channel: channel,
	}
}
