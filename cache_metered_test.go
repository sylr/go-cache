package cache

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe("localhost:9000", nil)
		if err != nil {
			panic(err)
		}
	}()
}

func TestMeteredCache(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)

	a, found := tc.Get("a")
	if found || a != nil {
		t.Error("Getting A found value that shouldn't exist:", a)
	}

	b, found := tc.Get("b")
	if found || b != nil {
		t.Error("Getting B found value that shouldn't exist:", b)
	}

	c, found := tc.Get("c")
	if found || c != nil {
		t.Error("Getting C found value that shouldn't exist:", c)
	}

	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", "b", DefaultExpiration)
	tc.Set("c", 3.5, DefaultExpiration)

	x, found := tc.Get("a")
	if !found {
		t.Error("a was not found while getting a2")
	}
	if x == nil {
		t.Error("x for a is nil")
	} else if a2 := x.(int); a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}

	x, found = tc.Get("b")
	if !found {
		t.Error("b was not found while getting b2")
	}
	if x == nil {
		t.Error("x for b is nil")
	} else if b2 := x.(string); b2+"B" != "bB" {
		t.Error("b2 (which should be b) plus B does not equal bB; value:", b2)
	}

	x, found = tc.Get("c")
	if !found {
		t.Error("c was not found while getting c2")
	}
	if x == nil {
		t.Error("x for c is nil")
	} else if c2 := x.(float64); c2+1.2 != 4.7 {
		t.Error("c2 (which should be 3.5) plus 1.2 does not equal 4.7; value:", c2)
	}
}

func TestMeteredCacheTimes(t *testing.T) {
	var found bool

	tc := NewMetered(50*time.Millisecond, 1*time.Millisecond)
	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", 2, NoExpiration)
	tc.Set("c", 3, 20*time.Millisecond)
	tc.Set("d", 4, 70*time.Millisecond)

	<-time.After(25 * time.Millisecond)
	_, found = tc.Get("c")
	if found {
		t.Error("Found c when it should have been automatically deleted")
	}

	<-time.After(30 * time.Millisecond)
	_, found = tc.Get("a")
	if found {
		t.Error("Found a when it should have been automatically deleted")
	}

	_, found = tc.Get("b")
	if !found {
		t.Error("Did not find b even though it was set to never expire")
	}

	_, found = tc.Get("d")
	if !found {
		t.Error("Did not find d even though it was set to expire later than the default")
	}

	<-time.After(20 * time.Millisecond)
	_, found = tc.Get("d")
	if found {
		t.Error("Found d when it should have been automatically deleted (later than the default)")
	}
}

func TestMeteredNewFrom(t *testing.T) {
	m := map[string]Item{
		"a": Item{
			Object:     1,
			Expiration: 0,
		},
		"b": Item{
			Object:     2,
			Expiration: 0,
		},
	}
	tc := NewFrom(DefaultExpiration, 0, m)
	a, found := tc.Get("a")
	if !found {
		t.Fatal("Did not find a")
	}
	if a.(int) != 1 {
		t.Fatal("a is not 1")
	}
	b, found := tc.Get("b")
	if !found {
		t.Fatal("Did not find b")
	}
	if b.(int) != 2 {
		t.Fatal("b is not 2")
	}
}

func TestMeteredStorePointerToStruct(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("foo", &TestStruct{Num: 1}, DefaultExpiration)
	x, found := tc.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo")
	}
	foo := x.(*TestStruct)
	foo.Num++

	y, found := tc.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo (second time)")
	}
	bar := y.(*TestStruct)
	if bar.Num != 2 {
		t.Fatal("TestStruct.Num is not 2")
	}
}

func TestMeteredIncrementWithInt(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tint", 1, DefaultExpiration)
	err := tc.Increment("tint", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint")
	if !found {
		t.Error("tint was not found")
	}
	if x.(int) != 3 {
		t.Error("tint is not 3:", x)
	}
}

func TestMeteredIncrementWithInt8(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tint8", int8(1), DefaultExpiration)
	err := tc.Increment("tint8", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint8")
	if !found {
		t.Error("tint8 was not found")
	}
	if x.(int8) != 3 {
		t.Error("tint8 is not 3:", x)
	}
}

func TestMeteredIncrementWithInt16(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tint16", int16(1), DefaultExpiration)
	err := tc.Increment("tint16", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint16")
	if !found {
		t.Error("tint16 was not found")
	}
	if x.(int16) != 3 {
		t.Error("tint16 is not 3:", x)
	}
}

func TestMeteredIncrementWithInt32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tint32", int32(1), DefaultExpiration)
	err := tc.Increment("tint32", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint32")
	if !found {
		t.Error("tint32 was not found")
	}
	if x.(int32) != 3 {
		t.Error("tint32 is not 3:", x)
	}
}

func TestMeteredIncrementWithInt64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tint64", int64(1), DefaultExpiration)
	err := tc.Increment("tint64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint64")
	if !found {
		t.Error("tint64 was not found")
	}
	if x.(int64) != 3 {
		t.Error("tint64 is not 3:", x)
	}
}

func TestMeteredIncrementWithUint(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tuint", uint(1), DefaultExpiration)
	err := tc.Increment("tuint", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tuint")
	if !found {
		t.Error("tuint was not found")
	}
	if x.(uint) != 3 {
		t.Error("tuint is not 3:", x)
	}
}

func TestMeteredIncrementWithUintptr(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tuintptr", uintptr(1), DefaultExpiration)
	err := tc.Increment("tuintptr", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}

	x, found := tc.Get("tuintptr")
	if !found {
		t.Error("tuintptr was not found")
	}
	if x.(uintptr) != 3 {
		t.Error("tuintptr is not 3:", x)
	}
}

func TestMeteredIncrementWithUint8(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tuint8", uint8(1), DefaultExpiration)
	err := tc.Increment("tuint8", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tuint8")
	if !found {
		t.Error("tuint8 was not found")
	}
	if x.(uint8) != 3 {
		t.Error("tuint8 is not 3:", x)
	}
}

func TestMeteredIncrementWithUint16(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tuint16", uint16(1), DefaultExpiration)
	err := tc.Increment("tuint16", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}

	x, found := tc.Get("tuint16")
	if !found {
		t.Error("tuint16 was not found")
	}
	if x.(uint16) != 3 {
		t.Error("tuint16 is not 3:", x)
	}
}

func TestMeteredIncrementWithUint32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tuint32", uint32(1), DefaultExpiration)
	err := tc.Increment("tuint32", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tuint32")
	if !found {
		t.Error("tuint32 was not found")
	}
	if x.(uint32) != 3 {
		t.Error("tuint32 is not 3:", x)
	}
}

func TestMeteredIncrementWithUint64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tuint64", uint64(1), DefaultExpiration)
	err := tc.Increment("tuint64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}

	x, found := tc.Get("tuint64")
	if !found {
		t.Error("tuint64 was not found")
	}
	if x.(uint64) != 3 {
		t.Error("tuint64 is not 3:", x)
	}
}

func TestMeteredIncrementWithFloat32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("float32", float32(1.5), DefaultExpiration)
	err := tc.Increment("float32", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("float32")
	if !found {
		t.Error("float32 was not found")
	}
	if x.(float32) != 3.5 {
		t.Error("float32 is not 3.5:", x)
	}
}

func TestMeteredIncrementWithFloat64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("float64", float64(1.5), DefaultExpiration)
	err := tc.Increment("float64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("float64")
	if !found {
		t.Error("float64 was not found")
	}
	if x.(float64) != 3.5 {
		t.Error("float64 is not 3.5:", x)
	}
}

func TestMeteredIncrementFloatWithFloat32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("float32", float32(1.5), DefaultExpiration)
	err := tc.IncrementFloat("float32", 2)
	if err != nil {
		t.Error("Error incrementfloating:", err)
	}
	x, found := tc.Get("float32")
	if !found {
		t.Error("float32 was not found")
	}
	if x.(float32) != 3.5 {
		t.Error("float32 is not 3.5:", x)
	}
}

func TestMeteredIncrementFloatWithFloat64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("float64", float64(1.5), DefaultExpiration)
	err := tc.IncrementFloat("float64", 2)
	if err != nil {
		t.Error("Error incrementfloating:", err)
	}
	x, found := tc.Get("float64")
	if !found {
		t.Error("float64 was not found")
	}
	if x.(float64) != 3.5 {
		t.Error("float64 is not 3.5:", x)
	}
}

func TestMeteredDecrementWithInt(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("int", int(5), DefaultExpiration)
	err := tc.Decrement("int", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int")
	if !found {
		t.Error("int was not found")
	}
	if x.(int) != 3 {
		t.Error("int is not 3:", x)
	}
}

func TestMeteredDecrementWithInt8(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("int8", int8(5), DefaultExpiration)
	err := tc.Decrement("int8", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int8")
	if !found {
		t.Error("int8 was not found")
	}
	if x.(int8) != 3 {
		t.Error("int8 is not 3:", x)
	}
}

func TestMeteredDecrementWithInt16(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("int16", int16(5), DefaultExpiration)
	err := tc.Decrement("int16", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int16")
	if !found {
		t.Error("int16 was not found")
	}
	if x.(int16) != 3 {
		t.Error("int16 is not 3:", x)
	}
}

func TestMeteredDecrementWithInt32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("int32", int32(5), DefaultExpiration)
	err := tc.Decrement("int32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int32")
	if !found {
		t.Error("int32 was not found")
	}
	if x.(int32) != 3 {
		t.Error("int32 is not 3:", x)
	}
}

func TestMeteredDecrementWithInt64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("int64", int64(5), DefaultExpiration)
	err := tc.Decrement("int64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int64")
	if !found {
		t.Error("int64 was not found")
	}
	if x.(int64) != 3 {
		t.Error("int64 is not 3:", x)
	}
}

func TestMeteredDecrementWithUint(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uint", uint(5), DefaultExpiration)
	err := tc.Decrement("uint", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint")
	if !found {
		t.Error("uint was not found")
	}
	if x.(uint) != 3 {
		t.Error("uint is not 3:", x)
	}
}

func TestMeteredDecrementWithUintptr(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uintptr", uintptr(5), DefaultExpiration)
	err := tc.Decrement("uintptr", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uintptr")
	if !found {
		t.Error("uintptr was not found")
	}
	if x.(uintptr) != 3 {
		t.Error("uintptr is not 3:", x)
	}
}

func TestMeteredDecrementWithUint8(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uint8", uint8(5), DefaultExpiration)
	err := tc.Decrement("uint8", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint8")
	if !found {
		t.Error("uint8 was not found")
	}
	if x.(uint8) != 3 {
		t.Error("uint8 is not 3:", x)
	}
}

func TestMeteredDecrementWithUint16(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uint16", uint16(5), DefaultExpiration)
	err := tc.Decrement("uint16", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint16")
	if !found {
		t.Error("uint16 was not found")
	}
	if x.(uint16) != 3 {
		t.Error("uint16 is not 3:", x)
	}
}

func TestMeteredDecrementWithUint32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uint32", uint32(5), DefaultExpiration)
	err := tc.Decrement("uint32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint32")
	if !found {
		t.Error("uint32 was not found")
	}
	if x.(uint32) != 3 {
		t.Error("uint32 is not 3:", x)
	}
}

func TestMeteredDecrementWithUint64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uint64", uint64(5), DefaultExpiration)
	err := tc.Decrement("uint64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint64")
	if !found {
		t.Error("uint64 was not found")
	}
	if x.(uint64) != 3 {
		t.Error("uint64 is not 3:", x)
	}
}

func TestMeteredDecrementWithFloat32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("float32", float32(5.5), DefaultExpiration)
	err := tc.Decrement("float32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("float32")
	if !found {
		t.Error("float32 was not found")
	}
	if x.(float32) != 3.5 {
		t.Error("float32 is not 3:", x)
	}
}

func TestMeteredDecrementWithFloat64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("float64", float64(5.5), DefaultExpiration)
	err := tc.Decrement("float64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("float64")
	if !found {
		t.Error("float64 was not found")
	}
	if x.(float64) != 3.5 {
		t.Error("float64 is not 3:", x)
	}
}

func TestMeteredDecrementFloatWithFloat32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("float32", float32(5.5), DefaultExpiration)
	err := tc.DecrementFloat("float32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("float32")
	if !found {
		t.Error("float32 was not found")
	}
	if x.(float32) != 3.5 {
		t.Error("float32 is not 3:", x)
	}
}

func TestMeteredDecrementFloatWithFloat64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("float64", float64(5.5), DefaultExpiration)
	err := tc.DecrementFloat("float64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("float64")
	if !found {
		t.Error("float64 was not found")
	}
	if x.(float64) != 3.5 {
		t.Error("float64 is not 3:", x)
	}
}

func TestMeteredIncrementInt(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tint", 1, DefaultExpiration)
	n, err := tc.IncrementInt("tint", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("tint")
	if !found {
		t.Error("tint was not found")
	}
	if x.(int) != 3 {
		t.Error("tint is not 3:", x)
	}
}

func TestMeteredIncrementInt8(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tint8", int8(1), DefaultExpiration)
	n, err := tc.IncrementInt8("tint8", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("tint8")
	if !found {
		t.Error("tint8 was not found")
	}
	if x.(int8) != 3 {
		t.Error("tint8 is not 3:", x)
	}
}

func TestMeteredIncrementInt16(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tint16", int16(1), DefaultExpiration)
	n, err := tc.IncrementInt16("tint16", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("tint16")
	if !found {
		t.Error("tint16 was not found")
	}
	if x.(int16) != 3 {
		t.Error("tint16 is not 3:", x)
	}
}

func TestMeteredIncrementInt32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tint32", int32(1), DefaultExpiration)
	n, err := tc.IncrementInt32("tint32", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("tint32")
	if !found {
		t.Error("tint32 was not found")
	}
	if x.(int32) != 3 {
		t.Error("tint32 is not 3:", x)
	}
}

func TestMeteredIncrementInt64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tint64", int64(1), DefaultExpiration)
	n, err := tc.IncrementInt64("tint64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("tint64")
	if !found {
		t.Error("tint64 was not found")
	}
	if x.(int64) != 3 {
		t.Error("tint64 is not 3:", x)
	}
}

func TestMeteredIncrementUint(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tuint", uint(1), DefaultExpiration)
	n, err := tc.IncrementUint("tuint", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("tuint")
	if !found {
		t.Error("tuint was not found")
	}
	if x.(uint) != 3 {
		t.Error("tuint is not 3:", x)
	}
}

func TestMeteredIncrementUintptr(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tuintptr", uintptr(1), DefaultExpiration)
	n, err := tc.IncrementUintptr("tuintptr", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("tuintptr")
	if !found {
		t.Error("tuintptr was not found")
	}
	if x.(uintptr) != 3 {
		t.Error("tuintptr is not 3:", x)
	}
}

func TestMeteredIncrementUint8(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tuint8", uint8(1), DefaultExpiration)
	n, err := tc.IncrementUint8("tuint8", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("tuint8")
	if !found {
		t.Error("tuint8 was not found")
	}
	if x.(uint8) != 3 {
		t.Error("tuint8 is not 3:", x)
	}
}

func TestMeteredIncrementUint16(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tuint16", uint16(1), DefaultExpiration)
	n, err := tc.IncrementUint16("tuint16", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("tuint16")
	if !found {
		t.Error("tuint16 was not found")
	}
	if x.(uint16) != 3 {
		t.Error("tuint16 is not 3:", x)
	}
}

func TestMeteredIncrementUint32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tuint32", uint32(1), DefaultExpiration)
	n, err := tc.IncrementUint32("tuint32", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("tuint32")
	if !found {
		t.Error("tuint32 was not found")
	}
	if x.(uint32) != 3 {
		t.Error("tuint32 is not 3:", x)
	}
}

func TestMeteredIncrementUint64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("tuint64", uint64(1), DefaultExpiration)
	n, err := tc.IncrementUint64("tuint64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("tuint64")
	if !found {
		t.Error("tuint64 was not found")
	}
	if x.(uint64) != 3 {
		t.Error("tuint64 is not 3:", x)
	}
}

func TestMeteredIncrementFloat32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("float32", float32(1.5), DefaultExpiration)
	n, err := tc.IncrementFloat32("float32", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3.5 {
		t.Error("Returned number is not 3.5:", n)
	}
	x, found := tc.Get("float32")
	if !found {
		t.Error("float32 was not found")
	}
	if x.(float32) != 3.5 {
		t.Error("float32 is not 3.5:", x)
	}
}

func TestMeteredIncrementFloat64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("float64", float64(1.5), DefaultExpiration)
	n, err := tc.IncrementFloat64("float64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	if n != 3.5 {
		t.Error("Returned number is not 3.5:", n)
	}
	x, found := tc.Get("float64")
	if !found {
		t.Error("float64 was not found")
	}
	if x.(float64) != 3.5 {
		t.Error("float64 is not 3.5:", x)
	}
}

func TestMeteredDecrementInt8(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("int8", int8(5), DefaultExpiration)
	n, err := tc.DecrementInt8("int8", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("int8")
	if !found {
		t.Error("int8 was not found")
	}
	if x.(int8) != 3 {
		t.Error("int8 is not 3:", x)
	}
}

func TestMeteredDecrementInt16(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("int16", int16(5), DefaultExpiration)
	n, err := tc.DecrementInt16("int16", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("int16")
	if !found {
		t.Error("int16 was not found")
	}
	if x.(int16) != 3 {
		t.Error("int16 is not 3:", x)
	}
}

func TestMeteredDecrementInt32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("int32", int32(5), DefaultExpiration)
	n, err := tc.DecrementInt32("int32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("int32")
	if !found {
		t.Error("int32 was not found")
	}
	if x.(int32) != 3 {
		t.Error("int32 is not 3:", x)
	}
}

func TestMeteredDecrementInt64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("int64", int64(5), DefaultExpiration)
	n, err := tc.DecrementInt64("int64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("int64")
	if !found {
		t.Error("int64 was not found")
	}
	if x.(int64) != 3 {
		t.Error("int64 is not 3:", x)
	}
}

func TestMeteredDecrementUint(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uint", uint(5), DefaultExpiration)
	n, err := tc.DecrementUint("uint", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("uint")
	if !found {
		t.Error("uint was not found")
	}
	if x.(uint) != 3 {
		t.Error("uint is not 3:", x)
	}
}

func TestMeteredDecrementUintptr(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uintptr", uintptr(5), DefaultExpiration)
	n, err := tc.DecrementUintptr("uintptr", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("uintptr")
	if !found {
		t.Error("uintptr was not found")
	}
	if x.(uintptr) != 3 {
		t.Error("uintptr is not 3:", x)
	}
}

func TestMeteredDecrementUint8(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uint8", uint8(5), DefaultExpiration)
	n, err := tc.DecrementUint8("uint8", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("uint8")
	if !found {
		t.Error("uint8 was not found")
	}
	if x.(uint8) != 3 {
		t.Error("uint8 is not 3:", x)
	}
}

func TestMeteredDecrementUint16(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uint16", uint16(5), DefaultExpiration)
	n, err := tc.DecrementUint16("uint16", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("uint16")
	if !found {
		t.Error("uint16 was not found")
	}
	if x.(uint16) != 3 {
		t.Error("uint16 is not 3:", x)
	}
}

func TestMeteredDecrementUint32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uint32", uint32(5), DefaultExpiration)
	n, err := tc.DecrementUint32("uint32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("uint32")
	if !found {
		t.Error("uint32 was not found")
	}
	if x.(uint32) != 3 {
		t.Error("uint32 is not 3:", x)
	}
}

func TestMeteredDecrementUint64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uint64", uint64(5), DefaultExpiration)
	n, err := tc.DecrementUint64("uint64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("uint64")
	if !found {
		t.Error("uint64 was not found")
	}
	if x.(uint64) != 3 {
		t.Error("uint64 is not 3:", x)
	}
}

func TestMeteredDecrementFloat32(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("float32", float32(5), DefaultExpiration)
	n, err := tc.DecrementFloat32("float32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("float32")
	if !found {
		t.Error("float32 was not found")
	}
	if x.(float32) != 3 {
		t.Error("float32 is not 3:", x)
	}
}

func TestMeteredDecrementFloat64(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("float64", float64(5), DefaultExpiration)
	n, err := tc.DecrementFloat64("float64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	if n != 3 {
		t.Error("Returned number is not 3:", n)
	}
	x, found := tc.Get("float64")
	if !found {
		t.Error("float64 was not found")
	}
	if x.(float64) != 3 {
		t.Error("float64 is not 3:", x)
	}
}

func TestMeteredAdd(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	err := tc.Add("foo", "bar", DefaultExpiration)
	if err != nil {
		t.Error("Couldn't add foo even though it shouldn't exist")
	}
	err = tc.Add("foo", "baz", DefaultExpiration)
	if err == nil {
		t.Error("Successfully added another foo when it should have returned an error")
	}
}

func TestMeteredReplace(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	err := tc.Replace("foo", "bar", DefaultExpiration)
	if err == nil {
		t.Error("Replaced foo when it shouldn't exist")
	}
	tc.Set("foo", "bar", DefaultExpiration)
	err = tc.Replace("foo", "bar", DefaultExpiration)
	if err != nil {
		t.Error("Couldn't replace existing key foo")
	}
}

func TestMeteredDelete(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("foo", "bar", DefaultExpiration)
	tc.Delete("foo")
	x, found := tc.Get("foo")
	if found {
		t.Error("foo was found, but it should have been deleted")
	}
	if x != nil {
		t.Error("x is not nil:", x)
	}
}

func TestMeteredItemCount(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("foo", "1", DefaultExpiration)
	tc.Set("bar", "2", DefaultExpiration)
	tc.Set("baz", "3", DefaultExpiration)
	if n := tc.ItemCount(); n != 3 {
		t.Errorf("Item count is not 3: %d", n)
	}
}

func TestMeteredFlush(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("foo", "bar", DefaultExpiration)
	tc.Set("baz", "yes", DefaultExpiration)
	tc.Flush()
	x, found := tc.Get("foo")
	if found {
		t.Error("foo was found, but it should have been deleted")
	}
	if x != nil {
		t.Error("x is not nil:", x)
	}
	x, found = tc.Get("baz")
	if found {
		t.Error("baz was found, but it should have been deleted")
	}
	if x != nil {
		t.Error("x is not nil:", x)
	}
}

func TestMeteredIncrementOverflowInt(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("int8", int8(127), DefaultExpiration)
	err := tc.Increment("int8", 1)
	if err != nil {
		t.Error("Error incrementing int8:", err)
	}
	x, _ := tc.Get("int8")
	int8 := x.(int8)
	if int8 != -128 {
		t.Error("int8 did not overflow as expected; value:", int8)
	}

}

func TestMeteredIncrementOverflowUint(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uint8", uint8(255), DefaultExpiration)
	err := tc.Increment("uint8", 1)
	if err != nil {
		t.Error("Error incrementing int8:", err)
	}
	x, _ := tc.Get("uint8")
	uint8 := x.(uint8)
	if uint8 != 0 {
		t.Error("uint8 did not overflow as expected; value:", uint8)
	}
}

func TestMeteredDecrementUnderflowUint(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("uint8", uint8(0), DefaultExpiration)
	err := tc.Decrement("uint8", 1)
	if err != nil {
		t.Error("Error decrementing int8:", err)
	}
	x, _ := tc.Get("uint8")
	uint8 := x.(uint8)
	if uint8 != 255 {
		t.Error("uint8 did not underflow as expected; value:", uint8)
	}
}

func TestMeteredOnEvicted(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("foo", 3, DefaultExpiration)
	if tc.c.onEvicted != nil {
		t.Fatal("tc.onEvicted is not nil")
	}
	works := false
	tc.OnEvicted(func(k string, v interface{}) {
		if k == "foo" && v.(int) == 3 {
			works = true
		}
		tc.Set("bar", 4, DefaultExpiration)
	})
	tc.Delete("foo")
	x, _ := tc.Get("bar")
	if !works {
		t.Error("works bool not true")
	}
	if x.(int) != 4 {
		t.Error("bar was not 4")
	}
}

func TestMeteredCacheSerialization(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	testFillAndSerialize(t, tc)

	// Check if gob.Register behaves properly even after multiple gob.Register
	// on c.Items (many of which will be the same type)
	testFillAndSerialize(t, tc)
}

func TestMeteredFileSerialization(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	tc.Add("a", "a", DefaultExpiration)
	tc.Add("b", "b", DefaultExpiration)
	f, err := ioutil.TempFile("", "go-cache-cache.dat")
	if err != nil {
		t.Fatal("Couldn't create cache file:", err)
	}
	fname := f.Name()
	f.Close()
	tc.SaveFile(fname)

	oc := NewMetered(DefaultExpiration, 0)
	oc.Add("a", "aa", 0) // this should not be overwritten
	err = oc.LoadFile(fname)
	if err != nil {
		t.Error(err)
	}
	a, found := oc.Get("a")
	if !found {
		t.Error("a was not found")
	}
	astr := a.(string)
	if astr != "aa" {
		if astr == "a" {
			t.Error("a was overwritten")
		} else {
			t.Error("a is not aa")
		}
	}
	b, found := oc.Get("b")
	if !found {
		t.Error("b was not found")
	}
	if b.(string) != "b" {
		t.Error("b is not b")
	}
}

func TestMeteredSerializeUnserializable(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)
	ch := make(chan bool, 1)
	ch <- true
	tc.Set("chan", ch, DefaultExpiration)
	fp := &bytes.Buffer{}
	err := tc.Save(fp) // this should fail gracefully
	if err.Error() != "gob NewTypeObject can't handle type: chan bool" {
		t.Error("Error from Save was not gob NewTypeObject can't handle type chan bool:", err)
	}
}

func BenchmarkMeteredCacheGetExpiring(b *testing.B) {
	benchmarkMeteredCacheGet(b, 5*time.Minute)
}

func BenchmarkMeteredCacheGetNotExpiring(b *testing.B) {
	benchmarkMeteredCacheGet(b, NoExpiration)
}

func benchmarkMeteredCacheGet(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := NewMetered(exp, 0)
	tc.Set("foo", "bar", DefaultExpiration)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Get("foo")
	}
}

func BenchmarkMeteredRWMutexMapGet(b *testing.B) {
	b.StopTimer()
	m := map[string]string{
		"foo": "bar",
	}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_, _ = m["foo"] //nolint:gosimple
		mu.RUnlock()
	}
}

func BenchmarkMeteredRWMutexInterfaceMapGetStruct(b *testing.B) {
	b.StopTimer()
	s := struct{ name string }{name: "foo"}
	m := map[interface{}]string{
		s: "bar",
	}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_, _ = m[s] //nolint:gosimple
		mu.RUnlock()
	}
}

func BenchmarkMeteredRWMutexInterfaceMapGetString(b *testing.B) {
	b.StopTimer()
	m := map[interface{}]string{
		"foo": "bar",
	}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_, _ = m["foo"] //nolint:gosimple
		mu.RUnlock()
	}
}

func BenchmarkMeteredCacheGetConcurrentExpiring(b *testing.B) {
	benchmarkMeteredCacheGetConcurrent(b, 5*time.Minute)
}

func BenchmarkMeteredCacheGetConcurrentNotExpiring(b *testing.B) {
	benchmarkMeteredCacheGetConcurrent(b, NoExpiration)
}

func benchmarkMeteredCacheGetConcurrent(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := NewMetered(exp, 0)
	tc.Set("foo", "bar", DefaultExpiration)
	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < each; j++ {
				tc.Get("foo")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkMeteredRWMutexMapGetConcurrent(b *testing.B) {
	b.StopTimer()
	m := map[string]string{
		"foo": "bar",
	}
	mu := sync.RWMutex{}
	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < each; j++ {
				mu.RLock()
				_, _ = m["foo"] //nolint:gosimple
				mu.RUnlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkMeteredCacheGetManyConcurrentExpiring(b *testing.B) {
	benchmarkMeteredCacheGetManyConcurrent(b, 5*time.Minute)
}

func BenchmarkMeteredCacheGetManyConcurrentNotExpiring(b *testing.B) {
	benchmarkMeteredCacheGetManyConcurrent(b, NoExpiration)
}

func benchmarkMeteredCacheGetManyConcurrent(b *testing.B, exp time.Duration) {
	// This is the same as BenchmarkCacheGetConcurrent, but its result
	// can be compared against BenchmarkShardedCacheGetManyConcurrent
	// in sharded_test.go.
	b.StopTimer()
	n := 10000
	tc := NewMetered(exp, 0)
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		k := "foo" + strconv.Itoa(n)
		keys[i] = k
		tc.Set(k, "bar", DefaultExpiration)
	}
	each := b.N / n
	wg := new(sync.WaitGroup)
	wg.Add(n)
	for _, v := range keys {
		go func(k string) {
			for j := 0; j < each; j++ {
				tc.Get(k)
			}
			wg.Done()
		}(v)
	}
	b.StartTimer()
	wg.Wait()
}

func BenchmarkMeteredCacheSetExpiring(b *testing.B) {
	benchmarkMeteredCacheSet(b, 5*time.Minute)
}

func BenchmarkMeteredCacheSetNotExpiring(b *testing.B) {
	benchmarkMeteredCacheSet(b, NoExpiration)
}

func benchmarkMeteredCacheSet(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := NewMetered(exp, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Set("foo", "bar", DefaultExpiration)
	}
}

func BenchmarkMeteredRWMutexMapSet(b *testing.B) {
	b.StopTimer()
	m := map[string]string{}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		m["foo"] = "bar"
		mu.Unlock()
	}
}

func BenchmarkMeteredCacheSetDelete(b *testing.B) {
	b.StopTimer()
	tc := NewMetered(DefaultExpiration, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Set("foo", "bar", DefaultExpiration)
		tc.Delete("foo")
	}
}

func BenchmarkMeteredRWMutexMapSetDelete(b *testing.B) {
	b.StopTimer()
	m := map[string]string{}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		m["foo"] = "bar"
		mu.Unlock()
		mu.Lock()
		delete(m, "foo")
		mu.Unlock()
	}
}

func BenchmarkMeteredCacheSetDeleteSingleLock(b *testing.B) {
	b.StopTimer()
	tc := NewMetered(DefaultExpiration, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.c.mu.Lock()
		tc.c.set("foo", "bar", DefaultExpiration)
		tc.c.delete("foo")
		tc.c.mu.Unlock()
	}
}

func BenchmarkMeteredRWMutexMapSetDeleteSingleLock(b *testing.B) {
	b.StopTimer()
	m := map[string]string{}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		m["foo"] = "bar"
		delete(m, "foo")
		mu.Unlock()
	}
}

func BenchmarkMeteredIncrementInt(b *testing.B) {
	b.StopTimer()
	tc := NewMetered(DefaultExpiration, 0)
	tc.Set("foo", 0, DefaultExpiration)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.IncrementInt("foo", 1)
	}
}

func BenchmarkMeteredDeleteExpiredLoop(b *testing.B) {
	b.StopTimer()
	tc := NewMetered(5*time.Minute, 0)
	tc.c.mu.Lock()
	for i := 0; i < 100000; i++ {
		tc.c.set(strconv.Itoa(i), "bar", DefaultExpiration)
	}
	tc.c.mu.Unlock()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.DeleteExpired()
	}
}

func TestMeteredGetWithExpiration(t *testing.T) {
	tc := NewMetered(DefaultExpiration, 0)

	a, expiration, found := tc.GetWithExpiration("a")
	if found || a != nil || !expiration.IsZero() {
		t.Error("Getting A found value that shouldn't exist:", a)
	}

	b, expiration, found := tc.GetWithExpiration("b")
	if found || b != nil || !expiration.IsZero() {
		t.Error("Getting B found value that shouldn't exist:", b)
	}

	c, expiration, found := tc.GetWithExpiration("c")
	if found || c != nil || !expiration.IsZero() {
		t.Error("Getting C found value that shouldn't exist:", c)
	}

	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", "b", DefaultExpiration)
	tc.Set("c", 3.5, DefaultExpiration)
	tc.Set("d", 1, NoExpiration)
	tc.Set("e", 1, 50*time.Millisecond)

	x, expiration, found := tc.GetWithExpiration("a")
	if !found {
		t.Error("a was not found while getting a2")
	}
	if x == nil {
		t.Error("x for a is nil")
	} else if a2 := x.(int); a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for a is not a zeroed time")
	}

	x, expiration, found = tc.GetWithExpiration("b")
	if !found {
		t.Error("b was not found while getting b2")
	}
	if x == nil {
		t.Error("x for b is nil")
	} else if b2 := x.(string); b2+"B" != "bB" {
		t.Error("b2 (which should be b) plus B does not equal bB; value:", b2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for b is not a zeroed time")
	}

	x, expiration, found = tc.GetWithExpiration("c")
	if !found {
		t.Error("c was not found while getting c2")
	}
	if x == nil {
		t.Error("x for c is nil")
	} else if c2 := x.(float64); c2+1.2 != 4.7 {
		t.Error("c2 (which should be 3.5) plus 1.2 does not equal 4.7; value:", c2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for c is not a zeroed time")
	}

	x, expiration, found = tc.GetWithExpiration("d")
	if !found {
		t.Error("d was not found while getting d2")
	}
	if x == nil {
		t.Error("x for d is nil")
	} else if d2 := x.(int); d2+2 != 3 {
		t.Error("d (which should be 1) plus 2 does not equal 3; value:", d2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for d is not a zeroed time")
	}

	x, expiration, found = tc.GetWithExpiration("e")
	if !found {
		t.Error("e was not found while getting e2")
	}
	if x == nil {
		t.Error("x for e is nil")
	} else if e2 := x.(int); e2+2 != 3 {
		t.Error("e (which should be 1) plus 2 does not equal 3; value:", e2)
	}
	if expiration.UnixNano() != tc.c.items["e"].Expiration {
		t.Error("expiration for e is not the correct time")
	}
	if expiration.UnixNano() < time.Now().UnixNano() {
		t.Error("expiration for e is in the past")
	}
}
