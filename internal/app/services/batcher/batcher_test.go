package batcher_test

import (
	"fmt"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/app/services/batcher"
)

func makePayloads(n int) []app.Payload {
	payloads := make([]app.Payload, n)
	for i := range payloads {
		payloads[i] = app.Payload(fmt.Sprintf("payload-%d", i+1))
	}
	return payloads
}

func TestBatch(t *testing.T) {
	tests := []struct {
		name             string
		input            []app.Payload
		expect_count     int
		expect_per_batch []int
	}{
		{
			name:             "empty input",
			input:            makePayloads(0),
			expect_count:     1,
			expect_per_batch: []int{0},
		},
		{
			name:             "single event",
			input:            makePayloads(1),
			expect_count:     1,
			expect_per_batch: []int{1},
		},
		{
			name:             "exactly batch size",
			input:            makePayloads(5),
			expect_count:     1,
			expect_per_batch: []int{5},
		},
		{
			name:             "one over batch size",
			input:            makePayloads(6),
			expect_count:     2,
			expect_per_batch: []int{5, 1},
		},
		{
			name:             "exact multiple",
			input:            makePayloads(10),
			expect_count:     2,
			expect_per_batch: []int{5, 5},
		},
		{
			name:             "remainder",
			input:            makePayloads(11),
			expect_count:     3,
			expect_per_batch: []int{5, 5, 1},
		},
		{
			name:             "large input",
			input:            makePayloads(23),
			expect_count:     5,
			expect_per_batch: []int{5, 5, 5, 5, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := batcher.NewPayloadBatcherService()
			batches := svc.Batch(tt.input)

			if len(batches) != tt.expect_count {
				t.Errorf("batch count = %d, want %d", len(batches), tt.expect_count)
				return
			}

			for i, b := range batches {
				if len(b) != tt.expect_per_batch[i] {
					t.Errorf("batch[%d] len = %d, want %d", i, len(b), tt.expect_per_batch[i])
				}
			}

			var reassembled []app.Payload
			for _, b := range batches {
				reassembled = append(reassembled, b...)
			}
			if len(reassembled) != len(tt.input) {
				t.Errorf("total payloads after batching = %d, want %d", len(reassembled), len(tt.input))
			}
			for i, p := range reassembled {
				if string(p) != string(tt.input[i]) {
					t.Errorf("payload[%d] = %q, want %q", i, p, tt.input[i])
				}
			}
		})
	}
}
