package appsync_test

import (
	"context"
	"fmt"

	"github.com/exanubes/appsync"
)

func Example() {
	ctx := context.Background()

	client, err := appsync.Connect(ctx)
	defer client.Close(ctx)

	result, err := client.Subscribe(ctx, appsync.SubscribeCommandInput{})
	if err != nil {
		panic(err)
	}
	defer result.Sub.Close(ctx)

	var msg struct{ Text string }
	if err := result.Sub.Decode(ctx, &msg); err != nil {
		panic(err)
	}

	fmt.Println(msg.Text)
}
