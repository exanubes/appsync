package appsync

import "context"

// Connect establishes a new AppSync WebSocket connection and returns a Client.
func Connect() *AppsyncClient {
	return &AppsyncClient{}
}

type AppsyncClient struct{}

func (client *AppsyncClient) Publish(ctx context.Context, input PublishCommandInput) (*PublishCommandOutput, error) {
	return nil, nil
}

func (client *AppsyncClient) Subscribe(ctx context.Context, input SubscribeCommandInput) (*SubscribeCommandOutput, error) {
	return nil, nil
}

func (client *AppsyncClient) Close(ctx context.Context) error {
	return nil
}

func (client *AppsyncClient) Err(ctx context.Context, input PublishCommandInput) error {
	return nil
}
