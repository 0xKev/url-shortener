package memory_store

func NewInMemoryURLStore() *InMemoryURLStore {
	return &InMemoryURLStore{map[string]string{}}
}

type InMemoryURLStore struct {
	store map[string]string
}

func (i *InMemoryURLStore) GetExpandedURL(shortLink string) string {
	return i.store[shortLink]
}

func (i *InMemoryURLStore) RecordBaseURL(shortLink string) {
	i.store[shortLink] = shortLink
}
