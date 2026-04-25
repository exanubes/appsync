package subscription

import (
	"testing"
)

func create_sub() *Subscription {
	sub, _ := New("test", "test", 1)
	return sub
}

func TestConstructor(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Error("Expected registry, received nil")
	}
}

func TestRegister(t *testing.T) {
	registry := NewRegistry()
	sub := create_sub()
	registry.Register(sub)

	if len(registry.store) != 1 {
		t.Errorf("Expected 1 subscription in registry, received %d", len(registry.store))
	}
}

func TestGet(t *testing.T) {
	registry := NewRegistry()
	sub := create_sub()
	registry.Register(sub)
	result := registry.Get(sub.id)
	if result == nil {
		t.Error("Expected result to be subscription, received nil")
	}

	if result != sub {
		t.Error("Expected the subscriptions to be the same")
	}
}

func TestRemove(t *testing.T) {
	registry := NewRegistry()
	sub := create_sub()
	registry.Register(sub)
	registry.Remove(sub.id)

	result := registry.Get(sub.id)

	if result != nil {
		t.Error("Expected result to be nil")
	}
}
