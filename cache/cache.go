package cache

type Cache struct {
}

func NewCache(size int) *Cache {
	return &Cache{}
}

func (c *Cache) Set(key, value []byte, expire int) error {
	return nil
}

func (c *Cache) Get(key []byte) ([]byte, error) {
	return nil, nil
}

func (c *Cache) Touch(key []byte, expire int) error {
	return nil
}

func (c *Cache) Delete(key []byte) error {
	return nil
}
