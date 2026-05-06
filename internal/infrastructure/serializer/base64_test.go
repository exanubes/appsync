package serializer_test

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/exanubes/appsync/internal/app"
	"github.com/exanubes/appsync/internal/infrastructure/serializer"
)

func TestSerialize(t *testing.T) {
	tests := []struct {
		name          string
		input         app.Signature
		expect_result string
		expect_err    bool
	}{
		{
			name:          "serializes single-key signature",
			input:         app.Signature{"Authorization": "token-value"},
			expect_result: "eyJBdXRob3JpemF0aW9uIjoidG9rZW4tdmFsdWUifQ",
		},
		{
			name:          "serializes empty signature",
			input:         app.Signature{},
			expect_result: "e30",
		},
		{
			name:  "output round-trips to original signature",
			input: app.Signature{"host": "example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := serializer.New()
			got, err := s.Serialize(tt.input)

			if tt.expect_err && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expect_err && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expect_result != "" && got != tt.expect_result {
				t.Errorf("got %q, want %q", got, tt.expect_result)
			}
			if !tt.expect_err {
				decoded, err := base64.RawURLEncoding.DecodeString(got)
				if err != nil {
					t.Fatalf("result is not valid base64: %v", err)
				}
				var result app.Signature
				if err := json.Unmarshal(decoded, &result); err != nil {
					t.Fatalf("decoded result is not valid JSON: %v", err)
				}
				if !reflect.DeepEqual(result, tt.input) {
					t.Errorf("round-trip mismatch: got %v, want %v", result, tt.input)
				}
			}
		})
	}
}
