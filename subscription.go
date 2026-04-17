package appsync

import "context"

type ChannelSubscription struct{}

func (sub *ChannelSubscription) Close(ctx context.Context) error {
	return nil
}
func (sub *ChannelSubscription) Next(ctx context.Context) (*NextMessageOutput, error) {
	return nil, nil
}
func (sub *ChannelSubscription) Decode(ctx context.Context, value any) error {
	return nil
}
