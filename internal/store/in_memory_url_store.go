package memory_store

func NewInMemoryURLStore() *InMemoryURLStore {
	return &InMemoryURLStore{map[string]string{}}
}

type InMemoryURLStore struct {
	store map[string]string
}

func (i *InMemoryURLStore) Load(shortLink string) (string, bool) {
	baseURL, found := i.store[shortLink]
	return baseURL, found
}

func (i *InMemoryURLStore) Save(baseURL, shortLink string) {
	i.store[baseURL] = shortLink
}
