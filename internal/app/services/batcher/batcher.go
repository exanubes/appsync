package batcher

import "github.com/exanubes/appsync/internal/app"

type PayloadBatcher struct{}

func NewPayloadBatcherService() *PayloadBatcher {
	return &PayloadBatcher{}
}

var batch_size = 5

func (service *PayloadBatcher) Batch(events []app.Payload) []app.Batch {
	if len(events) <= batch_size {
		return []app.Batch{events}
	}
	var batches []app.Batch

	for start := 0; start < len(events); start += batch_size {
		end := min(start+batch_size, len(events))

		batches = append(batches, events[start:end])
	}

	return batches
}
