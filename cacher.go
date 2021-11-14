package cache

import (
	"constraints"
	"time"
)

type Numeric interface {
	constraints.Integer | constraints.Float
}

type AnyCacher[T any] interface {
	// Delete all expired items from the cache.
	DeleteExpired()
	// Add an item to the cache only if an item doesn't already exist for the given
	// key, or if the existing item has expired. Returns an error otherwise.
	Add(k string, x T, d time.Duration) error
	// Delete an item from the cache. Does nothing if the key is not in the cache.
	Delete(k string)
	// Delete all items from the cache.
	Flush()
	// Get an item from the cache. Returns the item or nil, and a bool indicating
	// whether the key was found.
	Get(k string) (T, bool)
	// GetWithExpiration returns an item and its expiration time from the cache.
	// It returns the item or nil, the expiration time if one is set (if the item
	// never expires a zero value for time.Time is returned), and a bool indicating
	// whether the key was found.
	GetWithExpiration(k string) (T, time.Time, bool)
	// ItemCount returns the number of items in the cache.
	ItemCount() int
	// Copies all unexpired items in the cache into a new map and returns it.
	Items() map[string]Item[T]
	// Set a new value for the cache key only if it already exists, and the existing
	// item hasn't expired. Returns an error otherwise.
	Replace(k string, x T, d time.Duration) error
	// Add an item to the cache, replacing any existing item, using the default
	// expiration.
	SetDefault(k string, x T)
	// Add an item to the cache, replacing any existing item. If the duration is 0
	// (DefaultExpiration), the cache's default expiration time is used. If it is -1
	// (NoExpiration), the item never expires.
	Set(k string, x T, d time.Duration)
	// Sets an (optional) function that is called with the key and value when an
	// item is evicted from the cache. (Including when it is deleted manually, but
	// not when it is overwritten.) Set to nil to disable.
	OnEvicted(f func(string, T))
}

type NumericCacher[T Numeric] interface {
	AnyCacher[T]
	// Decrement an item of type int8 by n. Returns an error if the item's value is
	// not an int8, or if it was not found. If there is no error, the decremented
	// value is returned.
	Decrement(k string, n T) (T, error)
	// Increment an item of type int32 by n. Returns an error if the item's value is
	// not an int32, or if it was not found. If there is no error, the incremented
	// value is returned.
	Increment(k string, n T) (T, error)
}
