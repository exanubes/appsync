package subscription

import "sync"

type Registry struct {
	store map[string]*Subscription
	mutex sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		store: make(map[string]*Subscription),
	}
}

func (registry *Registry) Register(sub *Subscription) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	registry.store[sub.id] = sub
}

func (registry *Registry) Remove(id string) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	delete(registry.store, id)
}

func (registry *Registry) Get(id string) *Subscription {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	sub, ok := registry.store[id]
	if ok {
		return sub
	}

	return nil
}

func (registry *Registry) Active() []string {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	ids := []string{}

	for id, sub := range registry.store {
		if sub.Active() {
			ids = append(ids, id)
		}
	}

	return ids
}
