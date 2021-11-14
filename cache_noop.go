package cache

import (
	"fmt"
	"time"
)

// Cache implements Cacher.
type NoopCache[T any] struct{}

// Set adds an item to the cache, replacing any existing item. If the duration is 0
// (DefaultExpiration), the cache's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (c *NoopCache[T]) Set(k string, x T, d time.Duration) {

}

// SetDefault adds an item to the cache, replacing any existing item, using the default
// expiration.
func (c *NoopCache[T]) SetDefault(k string, x T) {

}

// Add an item to the cache only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (c *NoopCache[T]) Add(k string, x T, d time.Duration) error {
	return nil
}

// Replace replaces a new value for the cache key only if it already exists, and the existing
// item hasn't expired. Returns an error otherwise.
func (c *NoopCache[T]) Replace(k string, x T, d time.Duration) error {
	return nil
}

// Get gets an item from the cache. Returns the item or nil, and a bool indicating
// whether the key was found.
func (c *NoopCache[T]) Get(k string) (T, bool) {
	var ret T
	return ret, false
}

// GetWithExpiration returns an item and its expiration time from the cache.
// It returns the item or nil, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func (c *NoopCache[T]) GetWithExpiration(k string) (T, time.Time, bool) {
	var ret T
	return ret, time.Time{}, false
}

// Cache implements Cacher
type NoopNumericCache[T Numeric] struct {
	NoopCache[T]
}

// Increment increments an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n. To retrieve the incremented value, use one
// of the specialized methods, e.g. IncrementInt64.
func (c *NoopNumericCache[T]) Increment(k string, n T) (T, error) {
	var ret T
	return ret, fmt.Errorf("Item not found")
}

// Decrement decrements an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n. To retrieve the decremented value, use one
// of the specialized methods, e.g. DecrementInt64.
func (c *NoopNumericCache[T]) Decrement(k string, n T) (T, error) {
	var ret T
	return ret, fmt.Errorf("Item not found")
}

// Delete deletes an item from the cache. Does nothing if the key is not in the cache.
func (c *NoopCache[T]) Delete(k string) {

}

// DeleteExpired deletes all expired items from the cache.
func (c *NoopCache[T]) DeleteExpired() {

}

// OnEvicted sets an (optional) function that is called with the key and value when an
// item is evicted from the cache. (Including when it is deleted manually, but
// not when it is overwritten.) Set to nil to disable.
func (c *NoopCache[T]) OnEvicted(f func(string, T)) {

}

// Items copies all unexpired items in the cache into a new map and returns it.
func (c *NoopCache[T]) Items() map[string]Item[T] {
	m := make(map[string]Item[T], 0)
	return m
}

// ItemCount returns the number of items in the cache. This may include items that have
// expired, but have not yet been cleaned up.
func (c *NoopCache[T]) ItemCount() int {
	return 0
}

// Flush Delete all items from the cache.
func (c *NoopCache[T]) Flush() {

}

func newNoopCache[T any]() *NoopCache[T] {
	return &NoopCache[T]{}
}

func newNoopNumericCache[T Numeric]() *NoopNumericCache[T] {
	return &NoopNumericCache[T]{}
}

// New returns a new NoopCache[T]
func NewNoop[T any](defaultExpiration, cleanupInterval time.Duration) *NoopCache[T] {
	return newNoopCache[T]()
}

// NewAnyCacher returns an AnyCacher[T] interface
func NewNoopAnyCacher[T any](defaultExpiration, cleanupInterval time.Duration) AnyCacher[T] {
	return NewNoop[T](defaultExpiration, cleanupInterval)
}

// NewAnyCacher returns a *NoopNumericCache[T]
func NewNoopNumeric[T Numeric](defaultExpiration, cleanupInterval time.Duration) *NoopNumericCache[T] {
	return newNoopNumericCache[T]()
}

// NewCacher returns a NumericCacher[T] interface
func NewNoopNumericCacher[T Numeric](defaultExpiration, cleanupInterval time.Duration) NumericCacher[T] {
	return NewNoopNumeric[T](defaultExpiration, cleanupInterval)
}

// NewNoopFrom returns a new *NoopCache[T] with a given default expiration duration
// and cleanup interval.
func NewNoopFrom[T any](defaultExpiration, cleanupInterval time.Duration, items map[string]Item[T]) *NoopCache[T] {
	return newNoopCache[T]()
}

// NewAnyCacherFrom returns a AnyCacher[T] interface.
func NewNoopAnyCacherFrom[T any](defaultExpiration, cleanupInterval time.Duration, items map[string]Item[T]) AnyCacher[T] {
	return NewNoopFrom(defaultExpiration, cleanupInterval, items)
}

// NewAnyCacherFrom returns a *NumericCache[T].
func NewNoopNumericFrom[T Numeric](defaultExpiration, cleanupInterval time.Duration, items map[string]Item[T]) *NoopNumericCache[T] {
	return newNoopNumericCache[T]()
}

// NewAnyCacherFrom returns a NumericCacher[T] interface.
func NewNoopNumericCacherFrom[T Numeric](defaultExpiration, cleanupInterval time.Duration, items map[string]Item[T]) NumericCacher[T] {
	return NewNoopNumericFrom(defaultExpiration, cleanupInterval, items)
}
