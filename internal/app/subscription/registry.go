package subscription

type Registry struct {
	store map[string]*Subscription
}

func NewRegistry() *Registry {
	return &Registry{
		store: make(map[string]*Subscription),
	}
}

func (registry *Registry) Register(sub *Subscription) {
	registry.store[sub.id] = sub
}

func (registry *Registry) Remove(id string) {
	delete(registry.store, id)
}

func (registry *Registry) Get(id string) *Subscription {
	sub, ok := registry.store[id]
	if ok {
		return sub
	}

	return nil
}
