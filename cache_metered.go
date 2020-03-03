package cache

import (
	"encoding/gob"
	"io"
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

// MeteredCache implements Cacher
type MeteredCache struct {
	janitor *meteredJanitor
	c       *cache
}

// Set adds an item to the cache, replacing any existing item. If the duration is 0
// (DefaultExpiration), the cache's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (mc *MeteredCache) Set(k string, x interface{}, d time.Duration) {
	mc.c.Set(k, x, d)

	cacheItem.Inc()
	cacheSetTotal.Inc()
}

// SetDefault adds an item to the cache, replacing any existing item, using the default
// expiration.
func (mc *MeteredCache) SetDefault(k string, x interface{}) {
	mc.Set(k, x, DefaultExpiration)
}

// Add an item to the cache only if an item doesn't already exist for the given
// key, or if the existing item has expired. Returns an error otherwise.
func (mc *MeteredCache) Add(k string, x interface{}, d time.Duration) error {
	err := mc.c.Add(k, x, d)
	if err == nil {
		cacheItem.Inc()
	}

	cacheAddTotal.Inc()
	return err
}

// Replace replaces a new value for the cache key only if it already exists, and the existing
// item hasn't expired. Returns an error otherwise.
func (mc *MeteredCache) Replace(k string, x interface{}, d time.Duration) error {
	err := mc.c.Replace(k, x, d)

	cacheReplaceTotal.Inc()
	return err
}

// Get gets an item from the cache. Returns the item or nil, and a bool indicating
// whether the key was found.
func (mc *MeteredCache) Get(k string) (interface{}, bool) {
	return mc.c.Get(k)
}

// GetWithExpiration returns an item and its expiration time from the cache.
// It returns the item or nil, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func (mc *MeteredCache) GetWithExpiration(k string) (interface{}, time.Time, bool) {
	return mc.c.GetWithExpiration(k)
}

// Increment increments an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n. To retrieve the incremented value, use one
// of the specialized methods, e.g. IncrementInt64.
func (mc *MeteredCache) Increment(k string, n int64) error {
	err := mc.c.Increment(k, n)

	cacheIncrementTotal.Inc()
	return err
}

// IncrementFloat increments an item of type float32 or float64 by n. Returns an error if the
// item's value is not floating point, if it was not found, or if it is not
// possible to increment it by n. Pass a negative number to decrement the
// value. To retrieve the incremented value, use one of the specialized methods,
// e.g. IncrementFloat64.
func (mc *MeteredCache) IncrementFloat(k string, n float64) error {
	err := mc.c.IncrementFloat(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return err
}

// IncrementInt increments an item of type int by n. Returns an error if the item's value is
// not an int, or if it was not found. If there is no error, the incremented
// value is returned.
func (mc *MeteredCache) IncrementInt(k string, n int) (int, error) {
	i, err := mc.c.IncrementInt(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// IncrementInt8 increments an item of type int8 by n. Returns an error if the item's value is
// not an int8, or if it was not found. If there is no error, the incremented
// value is returned.
func (mc *MeteredCache) IncrementInt8(k string, n int8) (int8, error) {
	i, err := mc.c.IncrementInt8(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// IncrementInt16 increments an item of type int16 by n. Returns an error if the item's value is
// not an int16, or if it was not found. If there is no error, the incremented
// value is returned.
func (mc *MeteredCache) IncrementInt16(k string, n int16) (int16, error) {
	i, err := mc.c.IncrementInt16(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// IncrementInt32 increments an item of type int32 by n. Returns an error if the item's value is
// not an int32, or if it was not found. If there is no error, the incremented
// value is returned.
func (mc *MeteredCache) IncrementInt32(k string, n int32) (int32, error) {
	i, err := mc.c.IncrementInt32(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// IncrementInt64 increments an item of type int64 by n. Returns an error if the item's value is
// not an int64, or if it was not found. If there is no error, the incremented
// value is returned.
func (mc *MeteredCache) IncrementInt64(k string, n int64) (int64, error) {
	i, err := mc.c.IncrementInt64(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// IncrementUint increments an item of type uint by n. Returns an error if the item's value is
// not an uint, or if it was not found. If there is no error, the incremented
// value is returned.
func (mc *MeteredCache) IncrementUint(k string, n uint) (uint, error) {
	i, err := mc.c.IncrementUint(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// IncrementUintptr increments an item of type uintptr by n. Returns an error if the item's value
// is not an uintptr, or if it was not found. If there is no error, the
// incremented value is returned.
func (mc *MeteredCache) IncrementUintptr(k string, n uintptr) (uintptr, error) {
	i, err := mc.c.IncrementUintptr(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// IncrementUint8 increments an item of type uint8 by n. Returns an error if the item's value
// is not an uint8, or if it was not found. If there is no error, the
// incremented value is returned.
func (mc *MeteredCache) IncrementUint8(k string, n uint8) (uint8, error) {
	i, err := mc.c.IncrementUint8(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// IncrementUint16 increments an item of type uint16 by n. Returns an error if the item's value
// is not an uint16, or if it was not found. If there is no error, the
// incremented value is returned.
func (mc *MeteredCache) IncrementUint16(k string, n uint16) (uint16, error) {
	i, err := mc.c.IncrementUint16(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// IncrementUint32 increments an item of type uint32 by n. Returns an error if the item's value
// is not an uint32, or if it was not found. If there is no error, the
// incremented value is returned.
func (mc *MeteredCache) IncrementUint32(k string, n uint32) (uint32, error) {
	i, err := mc.c.IncrementUint32(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// IncrementUint64 increments an item of type uint64 by n. Returns an error if the item's value
// is not an uint64, or if it was not found. If there is no error, the
// incremented value is returned.
func (mc *MeteredCache) IncrementUint64(k string, n uint64) (uint64, error) {
	i, err := mc.c.IncrementUint64(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// IncrementFloat32 increments an item of type float32 by n. Returns an error if the item's value
// is not an float32, or if it was not found. If there is no error, the
// incremented value is returned.
func (mc *MeteredCache) IncrementFloat32(k string, n float32) (float32, error) {
	i, err := mc.c.IncrementFloat32(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// IncrementFloat64 increments an item of type float64 by n. Returns an error if the item's value
// is not an float64, or if it was not found. If there is no error, the
// incremented value is returned.
func (mc *MeteredCache) IncrementFloat64(k string, n float64) (float64, error) {
	i, err := mc.c.IncrementFloat64(k, n)
	if err == nil {
		cacheIncrementTotal.Inc()
	}

	return i, err
}

// Decrement decrements an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n. To retrieve the decremented value, use one
// of the specialized methods, e.g. DecrementInt64.
func (mc *MeteredCache) Decrement(k string, n int64) error {
	err := mc.c.Decrement(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return err
}

// DecrementFloat decrements an item of type float32 or float64 by n. Returns an error if the
// item's value is not floating point, if it was not found, or if it is not
// possible to decrement it by n. Pass a negative number to decrement the
// value. To retrieve the decremented value, use one of the specialized methods,
// e.g. DecrementFloat64.
func (mc *MeteredCache) DecrementFloat(k string, n float64) error {
	err := mc.c.DecrementFloat(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return err
}

// DecrementInt decrements an item of type int by n. Returns an error if the item's value is
// not an int, or if it was not found. If there is no error, the decremented
// value is returned.
func (mc *MeteredCache) DecrementInt(k string, n int) (int, error) {
	nv, err := mc.c.DecrementInt(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// DecrementInt8 decrements an item of type int8 by n. Returns an error if the item's value is
// not an int8, or if it was not found. If there is no error, the decremented
// value is returned.
func (mc *MeteredCache) DecrementInt8(k string, n int8) (int8, error) {
	nv, err := mc.c.DecrementInt8(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// DecrementInt16 decrements an item of type int16 by n. Returns an error if the item's value is
// not an int16, or if it was not found. If there is no error, the decremented
// value is returned.
func (mc *MeteredCache) DecrementInt16(k string, n int16) (int16, error) {
	nv, err := mc.c.DecrementInt16(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// DecrementInt32 decrements an item of type int32 by n. Returns an error if the item's value is
// not an int32, or if it was not found. If there is no error, the decremented
// value is returned.
func (mc *MeteredCache) DecrementInt32(k string, n int32) (int32, error) {
	nv, err := mc.c.DecrementInt32(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// DecrementInt64 decrements an item of type int64 by n. Returns an error if the item's value is
// not an int64, or if it was not found. If there is no error, the decremented
// value is returned.
func (mc *MeteredCache) DecrementInt64(k string, n int64) (int64, error) {
	nv, err := mc.c.DecrementInt64(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// DecrementUint decrements an item of type uint by n. Returns an error if the item's value is
// not an uint, or if it was not found. If there is no error, the decremented
// value is returned.
func (mc *MeteredCache) DecrementUint(k string, n uint) (uint, error) {
	nv, err := mc.c.DecrementUint(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// DecrementUintptr decrements an item of type uintptr by n. Returns an error if the item's value
// is not an uintptr, or if it was not found. If there is no error, the
// decremented value is returned.
func (mc *MeteredCache) DecrementUintptr(k string, n uintptr) (uintptr, error) {
	nv, err := mc.c.DecrementUintptr(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// DecrementUint8 decrements an item of type uint8 by n. Returns an error if the item's value is
// not an uint8, or if it was not found. If there is no error, the decremented
// value is returned.
func (mc *MeteredCache) DecrementUint8(k string, n uint8) (uint8, error) {
	nv, err := mc.c.DecrementUint8(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// DecrementUint16 decrements an item of type uint16 by n. Returns an error if the item's value
// is not an uint16, or if it was not found. If there is no error, the
// decremented value is returned.
func (mc *MeteredCache) DecrementUint16(k string, n uint16) (uint16, error) {
	nv, err := mc.c.DecrementUint16(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// DecrementUint32 decrements an item of type uint32 by n. Returns an error if the item's value
// is not an uint32, or if it was not found. If there is no error, the
// decremented value is returned.
func (mc *MeteredCache) DecrementUint32(k string, n uint32) (uint32, error) {
	nv, err := mc.c.DecrementUint32(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// DecrementUint64 decrements an item of type uint64 by n. Returns an error if the item's value
// is not an uint64, or if it was not found. If there is no error, the
// decremented value is returned.
func (mc *MeteredCache) DecrementUint64(k string, n uint64) (uint64, error) {
	nv, err := mc.c.DecrementUint64(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// DecrementFloat32 decrements an item of type float32 by n. Returns an error if the item's value
// is not an float32, or if it was not found. If there is no error, the
// decremented value is returned.
func (mc *MeteredCache) DecrementFloat32(k string, n float32) (float32, error) {
	nv, err := mc.c.DecrementFloat32(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// DecrementFloat64 decrements an item of type float64 by n. Returns an error if the item's value
// is not an float64, or if it was not found. If there is no error, the
// decremented value is returned.
func (mc *MeteredCache) DecrementFloat64(k string, n float64) (float64, error) {
	nv, err := mc.c.DecrementFloat64(k, n)
	if err == nil {
		cacheDecrementTotal.Inc()
	}

	return nv, err
}

// Delete deletes an item from the cache. Does nothing if the key is not in the cache.
func (mc *MeteredCache) Delete(k string) {
	mc.c.mu.Lock()
	v, evicted := mc.delete(k)
	mc.c.mu.Unlock()
	if mc.c.onEvicted != nil && evicted {
		mc.c.onEvicted(k, v)
	}
}

func (mc *MeteredCache) delete(k string) (interface{}, bool) {
	ret, found := mc.c.delete(k)

	if found {
		cacheItem.Dec()
		cacheDeleteTotal.Inc()
	}

	return ret, found
}

// DeleteExpired deletes all expired items from the cache.
func (mc *MeteredCache) DeleteExpired() {
	mc.c.DeleteExpired()
}

// OnEvicted sets an (optional) function that is called with the key and value when an
// item is evicted from the cache. (Including when it is deleted manually, but
// not when it is overwritten.) Set to nil to disable.
func (mc *MeteredCache) OnEvicted(f func(string, interface{})) {
	mc.c.OnEvicted(f)
}

// Save writes the cache's items (using Gob) to an io.Writer.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (mc *MeteredCache) Save(w io.Writer) (err error) {
	return mc.c.Save(w)
}

// SaveFile saves the cache's items to the given filename, creating the file if it
// doesn't exist, and overwriting it if it does.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (mc *MeteredCache) SaveFile(fname string) error {
	return mc.c.SaveFile(fname)
}

// Load adds (Gob-serialized) cache items from an io.Reader, excluding any items with
// keys that already exist (and haven't expired) in the current cache.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (mc *MeteredCache) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	items := map[string]Item{}
	err := dec.Decode(&items)
	if err == nil {
		mc.c.mu.Lock()
		defer mc.c.mu.Unlock()
		for k, v := range items {
			ov, found := mc.c.get(k)
			if !found {
				mc.c.items[k] = v

				if ov == nil {
					cacheItem.Inc()
				}
			}
		}
	}
	return err
}

// LoadFile loads and adds cache items from the given filename, excluding any items with
// keys that already exist in the current cache.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (mc *MeteredCache) LoadFile(fname string) error {
	return mc.c.LoadFile(fname)
}

// Items copies all unexpired items in the cache into a new map and returns it.
func (mc *MeteredCache) Items() map[string]Item {
	return mc.c.Items()
}

// ItemCount returns the number of items in the cache. This may include items that have
// expired, but have not yet been cleaned up.
func (mc *MeteredCache) ItemCount() int {
	return mc.c.ItemCount()
}

// Flush Delete all items from the cache.
func (mc *MeteredCache) Flush() {
	mc.c.Flush()
	cacheFlushTotal.Inc()
}

type meteredJanitor struct {
	*janitor
}

func (j *meteredJanitor) Run(c *cache) {
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

func stopMeteredJanitor(mc *MeteredCache) {
	mc.janitor.stop <- true
}

func runMeteredJanitor(mc *MeteredCache, ci time.Duration) {
	j := &meteredJanitor{
		janitor: &janitor{
			Interval: ci,
			stop:     make(chan bool),
		},
	}
	mc.janitor = j
	go j.Run(mc.c)
}

func newMeteredCache(de time.Duration, m map[string]Item) *cache {
	c := newCache(de, m)
	cacheItem.Add(float64(len(m)))
	return c
}

func newMeteredWithJanitor(de time.Duration, ci time.Duration, m map[string]Item) *MeteredCache {
	c := newMeteredCache(de, m)
	// This trick ensures that the janitor goroutine (which--granted it
	// was enabled--is running DeleteExpired on mc forever) does not keep
	// the returned MC object from being garbage collected. When it is
	// garbage collected, the finalizer stops the janitor goroutine, after
	// which mc can be collected.
	MC := &MeteredCache{c: c}
	if ci > 0 {
		runMeteredJanitor(MC, ci)
		runtime.SetFinalizer(MC, stopMeteredJanitor)
	}
	return MC
}

// NewMetered returns a new cache with a given default expiration duration and cleanup
// interval. If the expiration duration is less than one (or NoExpiration),
// the items in the cache never expire (by default), and must be deleted
// manually. If the cleanup interval is less than one, expired items are not
// deleted from the cache before calling c.DeleteExpired().
func NewMetered(defaultExpiration, cleanupInterval time.Duration) *MeteredCache {
	items := make(map[string]Item)
	return newMeteredWithJanitor(defaultExpiration, cleanupInterval, items)
}

// NewMeteredCacher returns a Cacher interface implementing MeteredCache
func NewMeteredCacher(defaultExpiration, cleanupInterval time.Duration) Cacher {
	return NewMetered(defaultExpiration, cleanupInterval)
}

// NewMeteredFrom returns a new cache with a given default expiration duration and cleanup
// interval. If the expiration duration is less than one (or NoExpiration),
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
func NewMeteredFrom(defaultExpiration, cleanupInterval time.Duration, items map[string]Item) *MeteredCache {
	return newMeteredWithJanitor(defaultExpiration, cleanupInterval, items)
}

// NewMeteredCacherFrom returns a Cacher interface implementing MeteredCache
func NewMeteredCacherFrom(defaultExpiration, cleanupInterval time.Duration, items map[string]Item) Cacher {
	return NewMeteredFrom(defaultExpiration, cleanupInterval, items)
}
