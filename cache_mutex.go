package cache

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Item ...
type Item[T any] struct {
	Object     T
	Expiration int64
}

// Expired returns true if the item has expired.
func (item *Item[T]) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

const (
	// NoExpiration for use with functions that take an expiration time.
	NoExpiration time.Duration = -1
	// DefaultExpiration for use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to New() or
	// NewFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 0
)

// Cache implements Cacher
type Cache[T any] struct {
	*cache[T]
	// If this is confusing, see the comment at the bottom of New()
}

type cache[T any] struct {
	defaultExpiration time.Duration
	items             map[string]Item[T]
	mu                sync.RWMutex
	onEvicted         func(string, *T)
	janitor           *janitor[T]
}

// Set adds an item to the cache, replacing any existing item. If the duration is 0
// (DefaultExpiration), the cache's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (c *cache[T]) Set(k string, x T, d time.Duration) {
	// "Inlining" of set
	var e int64

	if d == DefaultExpiration {
		d = c.defaultExpiration
	}

	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[k] = Item[T]{
		Object:     x,
		Expiration: e,
	}
}

func (c *cache[T]) set(k string, x T, d time.Duration) {
	var e int64

	if d == DefaultExpiration {
		d = c.defaultExpiration
	}

	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}

	c.items[k] = Item[T]{
		Object:     x,
		Expiration: e,
	}
}

// SetDefault adds an item to the cache, replacing any existing item, using the default
// expiration.
func (c *cache[T]) SetDefault(k string, x T) {
	c.Set(k, x, DefaultExpiration)
}

// Add an item to the cache only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (c *cache[T]) Add(k string, x T, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, found := c.get(k)
	if found {
		return fmt.Errorf("Item %s already exists", k)
	}
	c.set(k, x, d)
	return nil
}

// Replace replaces a new value for the cache key only if it already exists, and the existing
// item hasn't expired. Returns an error otherwise.
func (c *cache[T]) Replace(k string, x T, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, found := c.get(k)
	if !found {
		return fmt.Errorf("Item %s doesn't exist", k)
	}
	c.set(k, x, d)
	return nil
}

// Get gets an item from the cache. Returns the item or nil, and a bool indicating
// whether the key was found.
func (c *cache[T]) Get(k string) (*T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// "Inlining" of get and Expired
	item, found := c.items[k]
	if !found {
		return nil, false
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}

	return &item.Object, true
}

// GetWithExpiration returns an item and its expiration time from the cache.
// It returns the item or nil, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func (c *cache[T]) GetWithExpiration(k string) (*T, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// "Inlining" of get and Expired
	item, found := c.items[k]
	if !found {
		return nil, time.Time{}, false
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, time.Time{}, false
		}

		// Return the item and the expiration time
		return &item.Object, time.Unix(0, item.Expiration), true
	}

	// If expiration <= 0 (i.e. no expiration time set) then return the item
	// and a zeroed time.Time
	return &item.Object, time.Time{}, true
}

// get returns an item from the cache
// key found and item not expired => (value, true)
// key found and item expired     => (value, false)
// key not found                  => (nil, false)
func (c *cache[T]) get(k string) (*T, bool) {
	item, found := c.items[k]
	if !found {
		return nil, false
	}
	// "Inlining" of Expired
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return &item.Object, false
		}
	}
	return &item.Object, true
}

// Cache implements Cacher
type NumericCache[T Numeric] struct {
	*numericCache[T]
	// If this is confusing, see the comment at the bottom of New()
}

type numericCache[T Numeric] struct {
	*cache[T]
}

// Increment increments an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n. To retrieve the incremented value, use one
// of the specialized methods, e.g. IncrementInt64.
func (c *numericCache[T]) Increment(k string, n T) (*T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, found := c.items[k]

	if !found || v.Expired() {
		return nil, fmt.Errorf("Item %s not found", k)
	}

	nv := v.Object + n
	v.Object = nv
	c.items[k] = v

	return &nv, nil
}

// Decrement decrements an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n. To retrieve the decremented value, use one
// of the specialized methods, e.g. DecrementInt64.
func (c *numericCache[T]) Decrement(k string, n T) (*T, error) {
	// TODO: Implement Increment and Decrement more cleanly.
	// (Cannot do Increment(k, n*-1) for uints.)
	c.mu.Lock()
	defer c.mu.Unlock()

	v, found := c.items[k]

	if !found || v.Expired() {
		return nil, fmt.Errorf("Item not found")
	}

	nv := v.Object - n
	v.Object = nv
	c.items[k] = v

	return &nv, nil
}

// Delete deletes an item from the cache. Does nothing if the key is not in the cache.
func (c *cache[T]) Delete(k string) {
	c.mu.Lock()
	v, evicted := c.delete(k)
	c.mu.Unlock()

	if c.onEvicted != nil && evicted {
		c.onEvicted(k, v)
	}
}

func (c *cache[T]) delete(k string) (*T, bool) {
	var found = false
	var ret T

	if v, ok := c.items[k]; ok {
		found = true
		ret = v.Object
		delete(c.items, k)
	}

	return &ret, found
}

type keyAndValue[T any] struct {
	key   string
	value *T
}

// DeleteExpired deletes all expired items from the cache.
func (c *cache[T]) DeleteExpired() {
	var evictedItems []keyAndValue[T]
	now := time.Now().UnixNano()
	c.mu.Lock()
	for k, v := range c.items {
		// "Inlining" of expired
		if v.Expiration > 0 && now > v.Expiration {
			ov, evicted := c.delete(k)
			if c.onEvicted != nil && evicted {
				evictedItems = append(evictedItems, keyAndValue[T]{k, ov})
			}
		}
	}
	c.mu.Unlock()
	for _, v := range evictedItems {
		c.onEvicted(v.key, v.value)
	}
}

func (c *cache[T]) stopJanitor() {
	c.janitor.stop <- true
}

func (c *cache[T]) setJanitor(j *janitor[T]) {
	c.janitor = j
}

// OnEvicted sets an (optional) function that is called with the key and value when an
// item is evicted from the cache. (Including when it is deleted manually, but
// not when it is overwritten.) Set to nil to disable.
func (c *cache[T]) OnEvicted(f func(string, *T)) {
	c.mu.Lock()
	c.onEvicted = f
	c.mu.Unlock()
}

// Items copies all unexpired items in the cache into a new map and returns it.
func (c *cache[T]) Items() map[string]Item[T] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m := make(map[string]Item[T], len(c.items))
	now := time.Now().UnixNano()
	for k, v := range c.items {
		// "Inlining" of Expired
		if v.Expiration > 0 {
			if now > v.Expiration {
				continue
			}
		}
		m[k] = v
	}
	return m
}

// ItemCount returns the number of items in the cache. This may include items that have
// expired, but have not yet been cleaned up.
func (c *cache[T]) ItemCount() int {
	c.mu.RLock()
	n := len(c.items)
	c.mu.RUnlock()
	return n
}

// Flush Delete all items from the cache.
func (c *cache[T]) Flush() {
	c.mu.Lock()
	c.items = map[string]Item[T]{}
	c.mu.Unlock()
}

type janitor[T any] struct {
	Interval time.Duration
	stop     chan bool
}

func (j *janitor[T]) Run(c AnyCacher[T]) {
	ticker := time.NewTicker(j.Interval)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

type cacherWithJanitor[T any] interface {
	setJanitor(j *janitor[T])
	stopJanitor()
}

func stopJanitor[T any](c cacherWithJanitor[T]) {
	c.stopJanitor()
}

func runJanitor[T any](c cacherWithJanitor[T], ci time.Duration) {
	j := &janitor[T]{
		Interval: ci,
		stop:     make(chan bool),
	}
	c.setJanitor(j)
	go j.Run(c.(AnyCacher[T]))
}

func newCache[T any](de time.Duration, m map[string]Item[T]) *cache[T] {
	if de == 0 {
		de = -1
	}
	c := &cache[T]{
		defaultExpiration: de,
		items:             m,
	}
	return c
}

func newNumericCache[T Numeric](de time.Duration, m map[string]Item[T]) *numericCache[T] {
	if de == 0 {
		de = -1
	}

	c := &cache[T]{
		defaultExpiration: de,
		items:             m,
	}
	nc := &numericCache[T]{c}
	return nc
}

func newCacheWithJanitor[T any](de time.Duration, ci time.Duration, m map[string]Item[T]) *Cache[T] {
	c := newCache(de, m)
	// This trick ensures that the janitor goroutine (which--granted it
	// was enabled--is running DeleteExpired on c forever) does not keep
	// the returned C object from being garbage collected. When it is
	// garbage collected, the finalizer stops the janitor goroutine, after
	// which c can be collected.
	C := &Cache[T]{c}
	if ci > 0 {
		runJanitor[T](c, ci)
		runtime.SetFinalizer(C, stopJanitor[T])
	}
	return C
}

func newNumericCacheWithJanitor[T Numeric](de time.Duration, ci time.Duration, m map[string]Item[T]) *NumericCache[T] {
	c := newNumericCache(de, m)
	// This trick ensures that the janitor goroutine (which--granted it
	// was enabled--is running DeleteExpired on c forever) does not keep
	// the returned C object from being garbage collected. When it is
	// garbage collected, the finalizer stops the janitor goroutine, after
	// which c can be collected.
	C := &NumericCache[T]{c}

	if ci > 0 {
		runJanitor[T](c, ci)
		runtime.SetFinalizer(C, stopJanitor[T])
	}
	return C
}

// New returns a new Cache[T] with a given default expiration duration and cleanup
// interval. If the expiration duration is less than one (or NoExpiration),
// the items in the cache never expire (by default), and must be deleted
// manually. If the cleanup interval is less than one, expired items are not
// deleted from the cache before calling c.DeleteExpired().
func New[T any](defaultExpiration, cleanupInterval time.Duration) *Cache[T] {
	items := make(map[string]Item[T])
	return newCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}

// NewAnyCacher returns an AnyCacher[T] interface
func NewAnyCacher[T any](defaultExpiration, cleanupInterval time.Duration) AnyCacher[T] {
	return New[T](defaultExpiration, cleanupInterval)
}

// NewAnyCacher returns a *NumericCache[T]
func NewNumeric[T Numeric](defaultExpiration, cleanupInterval time.Duration) *NumericCache[T] {
	items := make(map[string]Item[T])
	return newNumericCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}

// NewCacher returns a NumericCacher[T] interface
func NewNumericCacher[T Numeric](defaultExpiration, cleanupInterval time.Duration) NumericCacher[T] {
	return NewNumeric[T](defaultExpiration, cleanupInterval)
}

// NewFrom returns a new *Cache[T] with a given default expiration duration and
// cleanup interval. If the expiration duration is less than one (or NoExpiration),
// the items in the cache never expire (by default), and must be deleted
// manually. If the cleanup interval is less than one, expired items are not
// deleted from the cache before calling c.DeleteExpired().
//
// NewFrom() also accepts an items map which will serve as the underlying map
// for the cache. This is useful for starting from a deserialized cache
// (serialized using e.g. gob.Encode() on c.Items()), or passing in e.g.
// make(map[string]Item, 500) to improve startup performance when the cache
// is expected to reach a certain minimum size.
//
// Only the cache's methods synchronize access to this map, so it is not
// recommended to keep any references to the map around after creating a cache.
// If need be, the map can be accessed at a later point using c.Items() (subject
// to the same caveat.)
//
// Note regarding serialization: When using e.g. gob, make sure to
// gob.Register() the individual types stored in the cache before encoding a
// map retrieved with c.Items(), and to register those same types before
// decoding a blob containing an items map.
func NewFrom[T any](defaultExpiration, cleanupInterval time.Duration, items map[string]Item[T]) *Cache[T] {
	return newCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}

// NewAnyCacherFrom returns a AnyCacher[T] interface
func NewAnyCacherFrom[T any](defaultExpiration, cleanupInterval time.Duration, items map[string]Item[T]) AnyCacher[T] {
	return NewFrom(defaultExpiration, cleanupInterval, items)
}

// NewAnyCacherFrom returns a *NumericCache[T]
func NewNumericFrom[T Numeric](defaultExpiration, cleanupInterval time.Duration, items map[string]Item[T]) *NumericCache[T] {
	return newNumericCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}

// NewAnyCacherFrom returns a NumericCacher[T] interface
func NewNumericCacherFrom[T Numeric](defaultExpiration, cleanupInterval time.Duration, items map[string]Item[T]) NumericCacher[T] {
	return NewNumericFrom(defaultExpiration, cleanupInterval, items)
}
