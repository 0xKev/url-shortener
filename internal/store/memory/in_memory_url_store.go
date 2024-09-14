package memory_store

import "sync"

func NewInMemoryURLStore() *InMemoryURLStore {
	return &InMemoryURLStore{
		map[string]string{},
		sync.Mutex{},
	}
}

type InMemoryURLStore struct {
	store map[string]string
	mu    sync.Mutex
}

func (i *InMemoryURLStore) Load(shortLink string) (string, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	baseURL, found := i.store[shortLink]
	return baseURL, found
}

func (i *InMemoryURLStore) Save(shortLink, baseURL string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.store[shortLink] = baseURL
}
