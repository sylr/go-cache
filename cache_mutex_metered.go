package cache

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	cacheItem = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "go",
			Subsystem: "cache",
			Name:      "set",
			Help:      "Current number of cached items",
		},
	)

	cacheAddTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "go",
			Subsystem: "cache",
			Name:      "add_total",
			Help:      "Total number of add operations",
		},
	)

	cacheDecrementTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "go",
			Subsystem: "cache",
			Name:      "decrement_total",
			Help:      "Total number of decrement operations",
		},
	)

	cacheDeleteTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "go",
			Subsystem: "cache",
			Name:      "delete_total",
			Help:      "Total number of delete operations",
		},
	)

	cacheFlushTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "go",
			Subsystem: "cache",
			Name:      "flush_total",
			Help:      "Total number of flush operations",
		},
	)

	cacheIncrementTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "go",
			Subsystem: "cache",
			Name:      "increment_total",
			Help:      "Total number of increment operations",
		},
	)

	cacheReplaceTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "go",
			Subsystem: "cache",
			Name:      "replace_total",
			Help:      "Total number of replace operations",
		},
	)

	cacheSetTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "go",
			Subsystem: "cache",
			Name:      "set_total",
			Help:      "Total number of set operations",
		},
	)

	cacheJanitorLastRun = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "go",
			Subsystem: "cache",
			Name:      "janitor_last_run",
			Help:      "Timestamp of last janitor run",
		},
	)
)

func init() {
	prometheus.MustRegister(cacheItem)
	prometheus.MustRegister(cacheAddTotal)
	prometheus.MustRegister(cacheDecrementTotal)
	prometheus.MustRegister(cacheDeleteTotal)
	prometheus.MustRegister(cacheFlushTotal)
	prometheus.MustRegister(cacheIncrementTotal)
	prometheus.MustRegister(cacheReplaceTotal)
	prometheus.MustRegister(cacheSetTotal)
	prometheus.MustRegister(cacheJanitorLastRun)
}

// AnyCache implements AnyCacher.
type AnyMeteredCache[T any] struct {
	*anyMeteredCache[T]
	// If this is confusing, see the comment at the bottom of New()
}

type anyMeteredCache[T any] struct {
	c *anyCache[T]
}

// Set adds an item to the cache, replacing any existing item. If the duration is 0
// (DefaultExpiration), the cache's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (mc *anyMeteredCache[T]) Set(k string, x T, d time.Duration) {
	mc.c.Set(k, x, d)

	cacheItem.Inc()
	cacheSetTotal.Inc()
}

// SetDefault adds an item to the cache, replacing any existing item, using the default
// expiration.
func (mc *anyMeteredCache[T]) SetDefault(k string, x T) {
	mc.Set(k, x, DefaultExpiration)
}

// Add an item to the cache only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (mc *anyMeteredCache[T]) Add(k string, x T, d time.Duration) error {
	defer func() {
		cacheItem.Inc()
		cacheSetTotal.Inc()
	}()

	return mc.c.Add(k, x, d)
}

// Replace replaces a new value for the cache key only if it already exists, and the existing
// item hasn't expired. Returns an error otherwise.
func (mc *anyMeteredCache[T]) Replace(k string, x T, d time.Duration) error {
	defer func() {
		cacheReplaceTotal.Inc()
	}()

	return mc.c.Replace(k, x, d)
}

// Get gets an item from the cache. Returns the item or nil, and a bool indicating
// whether the key was found.
func (mc *anyMeteredCache[T]) Get(k string) (T, bool) {
	return mc.c.Get(k)
}

// GetWithExpiration returns an item and its expiration time from the cache.
// It returns the item or nil, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func (mc *anyMeteredCache[T]) GetWithExpiration(k string) (T, time.Time, bool) {
	return mc.c.GetWithExpiration(k)
}

// NumericMeteredCache implements NumericCacher.
type NumericMeteredCache[T Numeric] struct {
	*numericMeteredCache[T]
	// If this is confusing, see the comment at the bottom of New()
}

type numericMeteredCache[T Numeric] struct {
	*numericCache[T]
}

// Increment increments an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n. To retrieve the incremented value, use one
// of the specialized methods, e.g. IncrementInt64.
func (nmc *numericMeteredCache[T]) Increment(k string, n T) (T, error) {
	defer cacheIncrementTotal.Inc()
	return nmc.Increment(k, n)
}

// Decrement decrements an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n. To retrieve the decremented value, use one
// of the specialized methods, e.g. DecrementInt64.
func (nmc *numericMeteredCache[T]) Decrement(k string, n T) (T, error) {
	defer cacheDecrementTotal.Inc()
	return nmc.Decrement(k, n)
}

// Delete deletes an item from the cache. Does nothing if the key is not in the cache.
func (mc *anyMeteredCache[T]) Delete(k string) {
	mc.c.mu.Lock()
	v, evicted := mc.delete(k)
	mc.c.mu.Unlock()
	if mc.c.onEvicted != nil && evicted {
		mc.c.onEvicted(k, v)
	}
}

func (mc *anyMeteredCache[T]) delete(k string) (T, bool) {
	ret, found := mc.c.delete(k)

	if found {
		cacheItem.Dec()
		cacheDeleteTotal.Inc()
	}

	return ret, found
}

// DeleteExpired deletes all expired items from the cache.
func (mc *anyMeteredCache[T]) DeleteExpired() {
	var evictedItems []keyAndValue[T]
	now := time.Now().UnixNano()
	mc.c.mu.Lock()
	for k, v := range mc.c.items {
		// "Inlining" of expired
		if v.Expiration > 0 && now > v.Expiration {
			ov, evicted := mc.delete(k)
			if mc.c.onEvicted != nil && evicted {
				evictedItems = append(evictedItems, keyAndValue[T]{k, ov})
			}
		}
	}
	mc.c.mu.Unlock()
	for _, v := range evictedItems {
		mc.c.onEvicted(v.key, v.value)
	}
}

func (mc *anyMeteredCache[T]) stopJanitor() {
	mc.c.janitor.stop <- true
}

func (mc *anyMeteredCache[T]) setJanitor(j *janitor[T]) {
	mc.c.janitor = j
}

// OnEvicted sets an (optional) function that is called with the key and value when an
// item is evicted from the cache. (Including when it is deleted manually, but
// not when it is overwritten.) Set to nil to disable.
func (mc *anyMeteredCache[T]) OnEvicted(f func(string, T)) {
	mc.c.mu.Lock()
	mc.c.onEvicted = f
	mc.c.mu.Unlock()
}

// Items copies all unexpired items in the cache into a new map and returns it.
func (mc *anyMeteredCache[T]) Items() map[string]Item[T] {
	return mc.c.Items()
}

// ItemCount returns the number of items in the cache. This may include items that have
// expired, but have not yet been cleaned up.
func (mc *anyMeteredCache[T]) ItemCount() int {
	return mc.c.ItemCount()
}

// Flush Delete all items from the cache.
func (mc *anyMeteredCache[T]) Flush() {
	mc.c.Flush()
}

type meteredJanitor[T any] struct {
	*janitor[T]
}

func (j *meteredJanitor[T]) Run(c AnyCacher[T]) {
	ticker := time.NewTicker(j.Interval)
	for {
		select {
		case <-ticker.C:
			cacheJanitorLastRun.Set(float64(time.Now().Unix()))
			c.DeleteExpired()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

type meteredCacherWithJanitor[T any] interface {
	setJanitor(j *meteredJanitor[T])
	stopJanitor()
}

func stopMeteredJanitor[T any](c meteredCacherWithJanitor[T]) {
	c.stopJanitor()
}

func runMeteredJanitor[T any](c meteredCacherWithJanitor[T], ci time.Duration) {
	j := &janitor[T]{
		Interval: ci,
		stop:     make(chan bool),
	}
	mj := &meteredJanitor[T]{j}
	c.setJanitor(mj)
	go j.Run(c.(AnyCacher[T]))
}

func newAnyMeteredCache[T any](de time.Duration, m map[string]Item[T]) *anyMeteredCache[T] {
	if de == 0 {
		de = -1
	}

	c := newAnyCache(de, m)
	mc := &anyMeteredCache[T]{c}

	return mc
}

func newNumericMeteredCache[T Numeric](de time.Duration, m map[string]Item[T]) *numericMeteredCache[T] {
	if de == 0 {
		de = -1
	}

	nc := newNumericCache(de, m)
	nmc := &numericMeteredCache[T]{nc}

	return nmc
}

func newAnyMeteredCacheWithJanitor[T any](de time.Duration, ci time.Duration, m map[string]Item[T]) *AnyMeteredCache[T] {
	c := newAnyMeteredCache(de, m)
	// This trick ensures that the janitor goroutine (which--granted it
	// was enabled--is running DeleteExpired on c forever) does not keep
	// the returned C object from being garbage collected. When it is
	// garbage collected, the finalizer stops the janitor goroutine, after
	// which c can be collected.
	C := &AnyMeteredCache[T]{c}

	if ci > 0 {
		runJanitor[T](c, ci)
		runtime.SetFinalizer(C, stopJanitor[T])
	}
	return C
}

func newNumericMeteredCacheWithJanitor[T Numeric](de time.Duration, ci time.Duration, m map[string]Item[T]) *NumericMeteredCache[T] {
	c := newNumericMeteredCache(de, m)
	// This trick ensures that the janitor goroutine (which--granted it
	// was enabled--is running DeleteExpired on c forever) does not keep
	// the returned C object from being garbage collected. When it is
	// garbage collected, the finalizer stops the janitor goroutine, after
	// which c can be collected.
	C := &NumericMeteredCache[T]{c}

	if ci > 0 {
		runJanitor[T](c, ci)
		runtime.SetFinalizer(C, stopJanitor[T])
	}
	return C
}

// NewAnyMetered[T any](...) returns a new cache with a given default expiration duration and cleanup
// interval. If the expiration duration is less than one (or NoExpiration),
// the items in the cache never expire (by default), and must be deleted
// manually. If the cleanup interval is less than one, expired items are not
// deleted from the cache before calling c.DeleteExpired().
func NewAnyMetered[T any](defaultExpiration, cleanupInterval time.Duration) *AnyMeteredCache[T] {
	items := make(map[string]Item[T])
	return newAnyMeteredCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}

// NewMeteredCacher[T any](...) returns an AnyCacher[T] interface.
func NewMeteredCacher[T any](defaultExpiration, cleanupInterval time.Duration) AnyCacher[T] {
	return NewAnyMetered[T](defaultExpiration, cleanupInterval)
}

// NewAnyMetered[T any](...) returns a new cache with a given default expiration duration and cleanup
// interval. If the expiration duration is less than one (or NoExpiration),
// the items in the cache never expire (by default), and must be deleted
// manually. If the cleanup interval is less than one, expired items are not
// deleted from the cache before calling c.DeleteExpired().
func NewNumericMetered[T Numeric](defaultExpiration, cleanupInterval time.Duration) *NumericMeteredCache[T] {
	items := make(map[string]Item[T])
	return newNumericMeteredCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}

// NewMeteredCacher[T any](...) returns an NumericCacher[T] interface.
func NewNumericMeteredCacher[T Numeric](defaultExpiration, cleanupInterval time.Duration) NumericCacher[T] {
	return NewNumericMetered[T](defaultExpiration, cleanupInterval)
}
