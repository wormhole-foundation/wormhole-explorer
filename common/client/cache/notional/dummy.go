package notional

// DummyNotionalCache is a dummy notional cache.
type DummyNotionalCache struct {
}

// NewDummyNotionalCache init a new dummy notional cache.
func NewDummyNotionalCache() *DummyNotionalCache {
	return &DummyNotionalCache{}
}

// Get get notional cache value.
func (c *DummyNotionalCache) Get(symbol string) (NotionalCacheField, error) {
	return NotionalCacheField{}, nil
}

// Close the dummy cache.
func (c *DummyNotionalCache) Close() error {
	return nil
}
