package cache

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
)

type TestStruct struct {
	Num      int
	Children []*TestStruct
}

func TestCache(t *testing.T) {
	tc := New[any](DefaultExpiration, 0)

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
	} else if a2 := (x).(int); a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}

	x, found = tc.Get("b")
	if !found {
		t.Error("b was not found while getting b2")
	}
	if x == nil {
		t.Error("x for b is nil")
	} else if b2 := (x).(string); b2+"B" != "bB" {
		t.Error("b2 (which should be b) plus B does not equal bB; value:", b2)
	}

	x, found = tc.Get("c")
	if !found {
		t.Error("c was not found while getting c2")
	}
	if x == nil {
		t.Error("x for c is nil")
	} else if c2 := (x).(float64); c2+1.2 != 4.7 {
		t.Error("c2 (which should be 3.5) plus 1.2 does not equal 4.7; value:", c2)
	}
}

func TestCacheTimes(t *testing.T) {
	var found bool

	tc := New[int](50*time.Millisecond, 1*time.Millisecond)
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

func TestNewFrom(t *testing.T) {
	m := map[string]Item[int]{
		"a": Item[int]{
			Object:     1,
			Expiration: 0,
		},
		"b": Item[int]{
			Object:     2,
			Expiration: 0,
		},
	}
	tc := NewFrom(DefaultExpiration, 0, m)
	a, found := tc.Get("a")
	if !found {
		t.Fatal("Did not find a")
	}
	if a != 1 {
		t.Fatal("a is not 1")
	}
	b, found := tc.Get("b")
	if !found {
		t.Fatal("Did not find b")
	}
	if b != 2 {
		t.Fatal("b is not 2")
	}
}

func TestStorePointerToStruct(t *testing.T) {
	tc := New[*TestStruct](DefaultExpiration, 0)
	tc.Set("foo", &TestStruct{Num: 1}, DefaultExpiration)
	x, found := tc.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo")
	}
	foo := x
	foo.Num++

	y, found := tc.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo (second time)")
	}
	bar := *y
	if bar.Num != 2 {
		t.Fatal("TestStruct.Num is not 2")
	}
}

func TestIncrementWithInt(t *testing.T) {
	tc := NewNumeric[int](DefaultExpiration, 0)
	tc.Set("tint", 1, DefaultExpiration)
	_, err := tc.Increment("tint", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint")
	if !found {
		t.Error("tint was not found")
	}
	if x != 3 {
		t.Error("tint is not 3:", x)
	}
}

func TestIncrementWithInt8(t *testing.T) {
	tc := NewNumeric[int8](DefaultExpiration, 0)
	tc.Set("tint8", int8(1), DefaultExpiration)
	_, err := tc.Increment("tint8", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint8")
	if !found {
		t.Error("tint8 was not found")
	}
	if x != 3 {
		t.Error("tint8 is not 3:", x)
	}
}

func TestIncrementWithInt16(t *testing.T) {
	tc := NewNumeric[int16](DefaultExpiration, 0)
	tc.Set("tint16", int16(1), DefaultExpiration)
	_, err := tc.Increment("tint16", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint16")
	if !found {
		t.Error("tint16 was not found")
	}
	if x != 3 {
		t.Error("tint16 is not 3:", x)
	}
}

func TestIncrementWithInt32(t *testing.T) {
	tc := NewNumeric[int32](DefaultExpiration, 0)
	tc.Set("tint32", int32(1), DefaultExpiration)
	_, err := tc.Increment("tint32", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint32")
	if !found {
		t.Error("tint32 was not found")
	}
	if x != 3 {
		t.Error("tint32 is not 3:", x)
	}
}

func TestIncrementWithInt64(t *testing.T) {
	tc := NewNumeric[int64](DefaultExpiration, 0)
	tc.Set("tint64", int64(1), DefaultExpiration)
	_, err := tc.Increment("tint64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tint64")
	if !found {
		t.Error("tint64 was not found")
	}
	if x != 3 {
		t.Error("tint64 is not 3:", x)
	}
}

func TestIncrementWithUint(t *testing.T) {
	tc := NewNumeric[uint](DefaultExpiration, 0)
	tc.Set("tuint", uint(1), DefaultExpiration)
	_, err := tc.Increment("tuint", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tuint")
	if !found {
		t.Error("tuint was not found")
	}
	if x != 3 {
		t.Error("tuint is not 3:", x)
	}
}

func TestIncrementWithUintptr(t *testing.T) {
	tc := NewNumeric[uintptr](DefaultExpiration, 0)
	tc.Set("tuintptr", uintptr(1), DefaultExpiration)
	_, err := tc.Increment("tuintptr", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tuintptr")
	if !found {
		t.Error("tuintptr was not found")
	}
	if x != 3 {
		t.Error("tuintptr is not 3:", x)
	}
}

func TestIncrementWithUint8(t *testing.T) {
	tc := NewNumeric[uint8](DefaultExpiration, 0)
	tc.Set("tuint8", uint8(1), DefaultExpiration)
	_, err := tc.Increment("tuint8", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tuint8")
	if !found {
		t.Error("tuint8 was not found")
	}
	if x != 3 {
		t.Error("tuint8 is not 3:", x)
	}
}

func TestIncrementWithUint16(t *testing.T) {
	tc := NewNumeric[uint16](DefaultExpiration, 0)
	tc.Set("tuint16", uint16(1), DefaultExpiration)
	_, err := tc.Increment("tuint16", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tuint16")
	if !found {
		t.Error("tuint16 was not found")
	}
	if x != 3 {
		t.Error("tuint16 is not 3:", x)
	}
}

func TestIncrementWithUint32(t *testing.T) {
	tc := NewNumeric[uint32](DefaultExpiration, 0)
	tc.Set("tuint32", uint32(1), DefaultExpiration)
	_, err := tc.Increment("tuint32", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tuint32")
	if !found {
		t.Error("tuint32 was not found")
	}
	if x != 3 {
		t.Error("tuint32 is not 3:", x)
	}
}

func TestIncrementWithUint64(t *testing.T) {
	tc := NewNumeric[uint64](DefaultExpiration, 0)
	tc.Set("tuint64", uint64(1), DefaultExpiration)
	_, err := tc.Increment("tuint64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("tuint64")
	if !found {
		t.Error("tuint64 was not found")
	}
	if x != 3 {
		t.Error("tuint64 is not 3:", x)
	}
}

func TestIncrementWithFloat32(t *testing.T) {
	tc := NewNumeric[float32](DefaultExpiration, 0)
	tc.Set("float32", float32(1.5), DefaultExpiration)
	_, err := tc.Increment("float32", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("float32")
	if !found {
		t.Error("float32 was not found")
	}
	if x != 3.5 {
		t.Error("float32 is not 3.5:", x)
	}
}

func TestIncrementWithFloat64(t *testing.T) {
	tc := NewNumeric[float64](DefaultExpiration, 0)
	tc.Set("float64", float64(1.5), DefaultExpiration)
	_, err := tc.Increment("float64", 2)
	if err != nil {
		t.Error("Error incrementing:", err)
	}
	x, found := tc.Get("float64")
	if !found {
		t.Error("float64 was not found")
	}
	if x != 3.5 {
		t.Error("float64 is not 3.5:", x)
	}
}

func TestIncrementFloatWithFloat32(t *testing.T) {
	tc := NewNumeric[float32](DefaultExpiration, 0)
	tc.Set("float32", float32(1.5), DefaultExpiration)
	_, err := tc.Increment("float32", 2)
	if err != nil {
		t.Error("Error incrementfloating:", err)
	}
	x, found := tc.Get("float32")
	if !found {
		t.Error("float32 was not found")
	}
	if x != 3.5 {
		t.Error("float32 is not 3.5:", x)
	}
}

func TestIncrementFloatWithFloat64(t *testing.T) {
	tc := NewNumeric[float64](DefaultExpiration, 0)
	tc.Set("float64", float64(1.5), DefaultExpiration)
	_, err := tc.Increment("float64", 2)
	if err != nil {
		t.Error("Error incrementfloating:", err)
	}
	x, found := tc.Get("float64")
	if !found {
		t.Error("float64 was not found")
	}
	if x != 3.5 {
		t.Error("float64 is not 3.5:", x)
	}
}

func TestDecrementWithInt(t *testing.T) {
	tc := NewNumeric[int](DefaultExpiration, 0)
	tc.Set("int", int(5), DefaultExpiration)
	_, err := tc.Decrement("int", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int")
	if !found {
		t.Error("int was not found")
	}
	if x != 3 {
		t.Error("int is not 3:", x)
	}
}

func TestDecrementWithInt8(t *testing.T) {
	tc := NewNumeric[int8](DefaultExpiration, 0)
	tc.Set("int8", int8(5), DefaultExpiration)
	_, err := tc.Decrement("int8", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int8")
	if !found {
		t.Error("int8 was not found")
	}
	if x != 3 {
		t.Error("int8 is not 3:", x)
	}
}

func TestDecrementWithInt16(t *testing.T) {
	tc := NewNumeric[int16](DefaultExpiration, 0)
	tc.Set("int16", int16(5), DefaultExpiration)
	_, err := tc.Decrement("int16", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int16")
	if !found {
		t.Error("int16 was not found")
	}
	if x != 3 {
		t.Error("int16 is not 3:", x)
	}
}

func TestDecrementWithInt32(t *testing.T) {
	tc := NewNumeric[int32](DefaultExpiration, 0)
	tc.Set("int32", int32(5), DefaultExpiration)
	_, err := tc.Decrement("int32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int32")
	if !found {
		t.Error("int32 was not found")
	}
	if x != 3 {
		t.Error("int32 is not 3:", x)
	}
}

func TestDecrementWithInt64(t *testing.T) {
	tc := NewNumeric[int64](DefaultExpiration, 0)
	tc.Set("int64", int64(5), DefaultExpiration)
	_, err := tc.Decrement("int64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("int64")
	if !found {
		t.Error("int64 was not found")
	}
	if x != 3 {
		t.Error("int64 is not 3:", x)
	}
}

func TestDecrementWithUint(t *testing.T) {
	tc := NewNumeric[uint](DefaultExpiration, 0)
	tc.Set("uint", uint(5), DefaultExpiration)
	_, err := tc.Decrement("uint", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint")
	if !found {
		t.Error("uint was not found")
	}
	if x != 3 {
		t.Error("uint is not 3:", x)
	}
}

func TestDecrementWithUintptr(t *testing.T) {
	tc := NewNumeric[uintptr](DefaultExpiration, 0)
	tc.Set("uintptr", uintptr(5), DefaultExpiration)
	_, err := tc.Decrement("uintptr", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uintptr")
	if !found {
		t.Error("uintptr was not found")
	}
	if x != 3 {
		t.Error("uintptr is not 3:", x)
	}
}

func TestDecrementWithUint8(t *testing.T) {
	tc := NewNumeric[uint8](DefaultExpiration, 0)
	tc.Set("uint8", uint8(5), DefaultExpiration)
	_, err := tc.Decrement("uint8", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint8")
	if !found {
		t.Error("uint8 was not found")
	}
	if x != 3 {
		t.Error("uint8 is not 3:", x)
	}
}

func TestDecrementWithUint16(t *testing.T) {
	tc := NewNumeric[uint16](DefaultExpiration, 0)
	tc.Set("uint16", uint16(5), DefaultExpiration)
	_, err := tc.Decrement("uint16", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint16")
	if !found {
		t.Error("uint16 was not found")
	}
	if x != 3 {
		t.Error("uint16 is not 3:", x)
	}
}

func TestDecrementWithUint32(t *testing.T) {
	tc := NewNumeric[uint32](DefaultExpiration, 0)
	tc.Set("uint32", uint32(5), DefaultExpiration)
	_, err := tc.Decrement("uint32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint32")
	if !found {
		t.Error("uint32 was not found")
	}
	if x != 3 {
		t.Error("uint32 is not 3:", x)
	}
}

func TestDecrementWithUint64(t *testing.T) {
	tc := NewNumeric[uint64](DefaultExpiration, 0)
	tc.Set("uint64", uint64(5), DefaultExpiration)
	_, err := tc.Decrement("uint64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("uint64")
	if !found {
		t.Error("uint64 was not found")
	}
	if x != 3 {
		t.Error("uint64 is not 3:", x)
	}
}

func TestDecrementWithFloat32(t *testing.T) {
	tc := NewNumeric[float32](DefaultExpiration, 0)
	tc.Set("float32", float32(5.5), DefaultExpiration)
	_, err := tc.Decrement("float32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("float32")
	if !found {
		t.Error("float32 was not found")
	}
	if x != 3.5 {
		t.Error("float32 is not 3:", x)
	}
}

func TestDecrementWithFloat64(t *testing.T) {
	tc := NewNumeric[float64](DefaultExpiration, 0)
	tc.Set("float64", float64(5.5), DefaultExpiration)
	_, err := tc.Decrement("float64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("float64")
	if !found {
		t.Error("float64 was not found")
	}
	if x != 3.5 {
		t.Error("float64 is not 3:", x)
	}
}

func TestDecrementFloatWithFloat32(t *testing.T) {
	tc := NewNumeric[float32](DefaultExpiration, 0)
	tc.Set("float32", float32(5.5), DefaultExpiration)
	_, err := tc.Decrement("float32", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("float32")
	if !found {
		t.Error("float32 was not found")
	}
	if x != 3.5 {
		t.Error("float32 is not 3:", x)
	}
}

func TestDecrementFloatWithFloat64(t *testing.T) {
	tc := NewNumeric[float64](DefaultExpiration, 0)
	tc.Set("float64", float64(5.5), DefaultExpiration)
	_, err := tc.Decrement("float64", 2)
	if err != nil {
		t.Error("Error decrementing:", err)
	}
	x, found := tc.Get("float64")
	if !found {
		t.Error("float64 was not found")
	}
	if x != 3.5 {
		t.Error("float64 is not 3:", x)
	}
}

func TestIncrementInt(t *testing.T) {
	tc := NewNumeric[int](DefaultExpiration, 0)
	tc.Set("tint", 1, DefaultExpiration)
	n, err := tc.Increment("tint", 2)
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
	if x != 3 {
		t.Error("tint is not 3:", x)
	}
}

func TestIncrementInt8(t *testing.T) {
	tc := NewNumeric[int8](DefaultExpiration, 0)
	tc.Set("tint8", int8(1), DefaultExpiration)
	n, err := tc.Increment("tint8", 2)
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
	if x != 3 {
		t.Error("tint8 is not 3:", x)
	}
}

func TestIncrementInt16(t *testing.T) {
	tc := NewNumeric[int16](DefaultExpiration, 0)
	tc.Set("tint16", int16(1), DefaultExpiration)
	n, err := tc.Increment("tint16", 2)
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
	if x != 3 {
		t.Error("tint16 is not 3:", x)
	}
}

func TestIncrementInt32(t *testing.T) {
	tc := NewNumeric[int32](DefaultExpiration, 0)
	tc.Set("tint32", int32(1), DefaultExpiration)
	n, err := tc.Increment("tint32", 2)
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
	if x != 3 {
		t.Error("tint32 is not 3:", x)
	}
}

func TestIncrementInt64(t *testing.T) {
	tc := NewNumeric[int64](DefaultExpiration, 0)
	tc.Set("tint64", int64(1), DefaultExpiration)
	n, err := tc.Increment("tint64", 2)
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
	if x != 3 {
		t.Error("tint64 is not 3:", x)
	}
}

func TestIncrementUint(t *testing.T) {
	tc := NewNumeric[uint](DefaultExpiration, 0)
	tc.Set("tuint", uint(1), DefaultExpiration)
	n, err := tc.Increment("tuint", 2)
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
	if x != 3 {
		t.Error("tuint is not 3:", x)
	}
}

func TestIncrementUintptr(t *testing.T) {
	tc := NewNumeric[uintptr](DefaultExpiration, 0)
	tc.Set("tuintptr", uintptr(1), DefaultExpiration)
	n, err := tc.Increment("tuintptr", 2)
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
	if x != 3 {
		t.Error("tuintptr is not 3:", x)
	}
}

func TestIncrementUint8(t *testing.T) {
	tc := NewNumeric[uint8](DefaultExpiration, 0)
	tc.Set("tuint8", uint8(1), DefaultExpiration)
	n, err := tc.Increment("tuint8", 2)
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
	if x != 3 {
		t.Error("tuint8 is not 3:", x)
	}
}

func TestIncrementUint16(t *testing.T) {
	tc := NewNumeric[uint16](DefaultExpiration, 0)
	tc.Set("tuint16", uint16(1), DefaultExpiration)
	n, err := tc.Increment("tuint16", 2)
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
	if x != 3 {
		t.Error("tuint16 is not 3:", x)
	}
}

func TestIncrementUint32(t *testing.T) {
	tc := NewNumeric[uint32](DefaultExpiration, 0)
	tc.Set("tuint32", uint32(1), DefaultExpiration)
	n, err := tc.Increment("tuint32", 2)
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
	if x != 3 {
		t.Error("tuint32 is not 3:", x)
	}
}

func TestIncrementUint64(t *testing.T) {
	tc := NewNumeric[uint64](DefaultExpiration, 0)
	tc.Set("tuint64", uint64(1), DefaultExpiration)
	n, err := tc.Increment("tuint64", 2)
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
	if x != 3 {
		t.Error("tuint64 is not 3:", x)
	}
}

func TestIncrementFloat32(t *testing.T) {
	tc := NewNumeric[float32](DefaultExpiration, 0)
	tc.Set("float32", float32(1.5), DefaultExpiration)
	n, err := tc.Increment("float32", 2)
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
	if x != 3.5 {
		t.Error("float32 is not 3.5:", x)
	}
}

func TestIncrementFloat64(t *testing.T) {
	tc := NewNumeric[float64](DefaultExpiration, 0)
	tc.Set("float64", float64(1.5), DefaultExpiration)
	n, err := tc.Increment("float64", 2)
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
	if x != 3.5 {
		t.Error("float64 is not 3.5:", x)
	}
}

func TestDecrementInt8(t *testing.T) {
	tc := NewNumeric[int8](DefaultExpiration, 0)
	tc.Set("int8", int8(5), DefaultExpiration)
	n, err := tc.Decrement("int8", 2)
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
	if x != 3 {
		t.Error("int8 is not 3:", x)
	}
}

func TestDecrementInt16(t *testing.T) {
	tc := NewNumeric[int16](DefaultExpiration, 0)
	tc.Set("int16", int16(5), DefaultExpiration)
	n, err := tc.Decrement("int16", 2)
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
	if x != 3 {
		t.Error("int16 is not 3:", x)
	}
}

func TestDecrementInt32(t *testing.T) {
	tc := NewNumeric[int32](DefaultExpiration, 0)
	tc.Set("int32", int32(5), DefaultExpiration)
	n, err := tc.Decrement("int32", 2)
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
	if x != 3 {
		t.Error("int32 is not 3:", x)
	}
}

func TestDecrementInt64(t *testing.T) {
	tc := NewNumeric[int64](DefaultExpiration, 0)
	tc.Set("int64", int64(5), DefaultExpiration)
	n, err := tc.Decrement("int64", 2)
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
	if x != 3 {
		t.Error("int64 is not 3:", x)
	}
}

func TestDecrementUint(t *testing.T) {
	tc := NewNumeric[uint](DefaultExpiration, 0)
	tc.Set("uint", uint(5), DefaultExpiration)
	n, err := tc.Decrement("uint", 2)
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
	if x != 3 {
		t.Error("uint is not 3:", x)
	}
}

func TestDecrementUintptr(t *testing.T) {
	tc := NewNumeric[uintptr](DefaultExpiration, 0)
	tc.Set("uintptr", uintptr(5), DefaultExpiration)
	n, err := tc.Decrement("uintptr", 2)
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
	if x != 3 {
		t.Error("uintptr is not 3:", x)
	}
}

func TestDecrementUint8(t *testing.T) {
	tc := NewNumeric[uint8](DefaultExpiration, 0)
	tc.Set("uint8", uint8(5), DefaultExpiration)
	n, err := tc.Decrement("uint8", 2)
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
	if x != 3 {
		t.Error("uint8 is not 3:", x)
	}
}

func TestDecrementUint16(t *testing.T) {
	tc := NewNumeric[uint16](DefaultExpiration, 0)
	tc.Set("uint16", uint16(5), DefaultExpiration)
	n, err := tc.Decrement("uint16", 2)
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
	if x != 3 {
		t.Error("uint16 is not 3:", x)
	}
}

func TestDecrementUint32(t *testing.T) {
	tc := NewNumeric[uint32](DefaultExpiration, 0)
	tc.Set("uint32", uint32(5), DefaultExpiration)
	n, err := tc.Decrement("uint32", 2)
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
	if x != 3 {
		t.Error("uint32 is not 3:", x)
	}
}

func TestDecrementUint64(t *testing.T) {
	tc := NewNumeric[uint64](DefaultExpiration, 0)
	tc.Set("uint64", uint64(5), DefaultExpiration)
	n, err := tc.Decrement("uint64", 2)
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
	if x != 3 {
		t.Error("uint64 is not 3:", x)
	}
}

func TestDecrementFloat32(t *testing.T) {
	tc := NewNumeric[float32](DefaultExpiration, 0)
	tc.Set("float32", float32(5), DefaultExpiration)
	n, err := tc.Decrement("float32", 2)
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
	if x != 3 {
		t.Error("float32 is not 3:", x)
	}
}

func TestDecrementFloat64(t *testing.T) {
	tc := NewNumeric[float64](DefaultExpiration, 0)
	tc.Set("float64", float64(5), DefaultExpiration)
	n, err := tc.Decrement("float64", 2)
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
	if x != 3 {
		t.Error("float64 is not 3:", x)
	}
}

func TestAdd(t *testing.T) {
	tc := New[any](DefaultExpiration, 0)
	err := tc.Add("foo", "bar", DefaultExpiration)
	if err != nil {
		t.Error("Couldn't add foo even though it shouldn't exist")
	}
	err = tc.Add("foo", "baz", DefaultExpiration)
	if err == nil {
		t.Error("Successfully added another foo when it should have returned an error")
	}
}

func TestReplace(t *testing.T) {
	tc := New[any](DefaultExpiration, 0)
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

func TestDelete(t *testing.T) {
	tc := New[any](DefaultExpiration, 0)
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

func TestItemCount(t *testing.T) {
	tc := New[any](DefaultExpiration, 0)
	tc.Set("foo", "1", DefaultExpiration)
	tc.Set("bar", "2", DefaultExpiration)
	tc.Set("baz", "3", DefaultExpiration)
	if n := tc.ItemCount(); n != 3 {
		t.Errorf("Item count is not 3: %d", n)
	}
}

func TestFlush(t *testing.T) {
	tc := New[string](DefaultExpiration, 0)
	tc.Set("foo", "bar", DefaultExpiration)
	tc.Set("baz", "yes", DefaultExpiration)
	tc.Flush()
	x, found := tc.Get("foo")
	if found {
		t.Error("foo was found, but it should have been deleted")
	}
	if x != "" {
		t.Error("x is not empty:", x)
	}
	x, found = tc.Get("baz")
	if found {
		t.Error("baz was found, but it should have been deleted")
	}
	if x != "" {
		t.Error("x is not empty:", x)
	}
}

func TestIncrementOverflowInt(t *testing.T) {
	tc := NewNumeric[int8](DefaultExpiration, 0)
	tc.Set("int8", int8(127), DefaultExpiration)
	_, err := tc.Increment("int8", 1)
	if err != nil {
		t.Error("Error incrementing int8:", err)
	}
	x, _ := tc.Get("int8")
	if x != -128 {
		t.Error("int8 did not overflow as expected; value:", x)
	}
}

func TestIncrementOverflowUint(t *testing.T) {
	tc := NewNumeric[uint8](DefaultExpiration, 0)
	tc.Set("uint8", uint8(255), DefaultExpiration)
	_, err := tc.Increment("uint8", 1)
	if err != nil {
		t.Error("Error incrementing int8:", err)
	}
	x, _ := tc.Get("uint8")
	if x != 0 {
		t.Error("uint8 did not overflow as expected; value:", x)
	}
}

func TestDecrementUnderflowUint(t *testing.T) {
	tc := NewNumeric[uint8](DefaultExpiration, 0)
	tc.Set("uint8", uint8(0), DefaultExpiration)
	_, err := tc.Decrement("uint8", 1)
	if err != nil {
		t.Error("Error decrementing int8:", err)
	}
	x, _ := tc.Get("uint8")
	if x != 255 {
		t.Error("uint8 did not underflow as expected; value:", x)
	}
}

func TestOnEvicted(t *testing.T) {
	tc := NewNumeric[int](DefaultExpiration, 0)
	tc.Set("foo", 3, DefaultExpiration)
	if tc.onEvicted != nil {
		t.Fatal("tc.onEvicted is not nil")
	}
	works := false
	tc.OnEvicted(func(k string, v int) {
		if k == "foo" && v == 3 {
			works = true
		}
		tc.Set("bar", 4, DefaultExpiration)
	})
	tc.Delete("foo")
	x, _ := tc.Get("bar")
	if !works {
		t.Error("works bool not true")
	}
	if x != 4 {
		t.Error("bar was not 4")
	}
}

func TestFinalizerNew(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer runtime.GC() // Force gc before verifying there are no leaked goroutines

	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < 5+rand.Intn(15); i++ {
		// Scope the creation of the Cache so it gets deleted by the GC at the end
		// of the scope.
		func() {
			tc := New[string](time.Second, time.Second)
			tc.SetDefault("pwet", "pwet")
		}()
	}
}

func TestFinalizerNewAny(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer runtime.GC() // Force gc before verifying there are no leaked goroutines

	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < 5+rand.Intn(15); i++ {
		// Scope the creation of the Cache so it gets deleted by the GC at the end
		// of the scope.
		func() {
			tc := NewAny[string](time.Second, time.Second)
			tc.SetDefault("pwet", "pwet")
		}()
	}
}

func TestFinalizerNewAnyCacher(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer runtime.GC() // Force gc before verifying there are no leaked goroutines

	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < 5+rand.Intn(15); i++ {
		// Scope the creation of the Cache so it gets deleted by the GC at the end
		// of the scope.
		func() {
			tc := NewAnyCacher[string](time.Second, time.Second)
			tc.SetDefault("pwet", "pwet")
		}()
	}
}

func TestFinalizerNewNumeric(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer runtime.GC() // Force gc before verifying there are no leaked goroutines

	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < 5+rand.Intn(15); i++ {
		// Scope the creation of the Cache so it gets deleted by the GC at the end
		// of the scope.
		func() {
			tc := NewNumeric[int](time.Second, time.Second)
			tc.SetDefault("pwet", 42)
		}()
	}
}

func TestFinalizerNewNumericCacher(t *testing.T) {
	defer goleak.VerifyNone(t)
	defer runtime.GC() // Force gc before verifying there are no leaked goroutines

	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < 5+rand.Intn(15); i++ {
		// Scope the creation of the Cache so it gets deleted by the GC at the end
		// of the scope.
		func() {
			tc := NewNumericCacher[int](time.Second, time.Second)
			tc.SetDefault("pwet", 42)
		}()
	}
}

func BenchmarkCacheGetExpiring(b *testing.B) {
	benchmarkCacheGet(b, 5*time.Minute)
}

func BenchmarkCacheGetNotExpiring(b *testing.B) {
	benchmarkCacheGet(b, NoExpiration)
}

func benchmarkCacheGet(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := New[string](exp, 0)
	tc.Set("foo", "bar", DefaultExpiration)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Get("foo")
	}
}

func BenchmarkRWMutexMapGet(b *testing.B) {
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

func BenchmarkRWMutexInterfaceMapGetStruct(b *testing.B) {
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

func BenchmarkRWMutexInterfaceMapGetString(b *testing.B) {
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

func BenchmarkCacheGetConcurrentExpiring(b *testing.B) {
	benchmarkCacheGetConcurrent(b, 5*time.Minute)
}

func BenchmarkCacheGetConcurrentNotExpiring(b *testing.B) {
	benchmarkCacheGetConcurrent(b, NoExpiration)
}

func benchmarkCacheGetConcurrent(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := New[string](exp, 0)
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

func BenchmarkRWMutexMapGetConcurrent(b *testing.B) {
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

func BenchmarkCacheGetManyConcurrentExpiring(b *testing.B) {
	benchmarkCacheGetManyConcurrent(b, 5*time.Minute)
}

func BenchmarkCacheGetManyConcurrentNotExpiring(b *testing.B) {
	benchmarkCacheGetManyConcurrent(b, NoExpiration)
}

func benchmarkCacheGetManyConcurrent(b *testing.B, exp time.Duration) {
	// This is the same as BenchmarkCacheGetConcurrent, but its result
	// can be compared against BenchmarkShardedCacheGetManyConcurrent
	// in sharded_test.go.
	b.StopTimer()
	n := 10000
	tc := New[string](exp, 0)
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		k := "foo" + strconv.Itoa(i)
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

func BenchmarkCacheSetExpiring(b *testing.B) {
	benchmarkCacheSet(b, 5*time.Minute)
}

func BenchmarkCacheSetNotExpiring(b *testing.B) {
	benchmarkCacheSet(b, NoExpiration)
}

func benchmarkCacheSet(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := New[string](exp, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Set("foo", "bar", DefaultExpiration)
	}
}

func BenchmarkRWMutexMapSet(b *testing.B) {
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

func BenchmarkCacheSetDelete(b *testing.B) {
	b.StopTimer()
	tc := New[string](DefaultExpiration, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Set("foo", "bar", DefaultExpiration)
		tc.Delete("foo")
	}
}

func BenchmarkRWMutexMapSetDelete(b *testing.B) {
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

func BenchmarkCacheSetDeleteSingleLock(b *testing.B) {
	b.StopTimer()
	tc := New[string](DefaultExpiration, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.mu.Lock()
		tc.set("foo", "bar", DefaultExpiration)
		tc.delete("foo")
		tc.mu.Unlock()
	}
}

func BenchmarkRWMutexMapSetDeleteSingleLock(b *testing.B) {
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

func BenchmarkIncrementInt(b *testing.B) {
	b.StopTimer()
	tc := NewNumeric[int64](DefaultExpiration, 0)
	tc.Set("foo", 0, DefaultExpiration)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Increment("foo", 1)
	}
}

func BenchmarkDeleteExpiredLoop(b *testing.B) {
	b.StopTimer()
	tc := New[string](5*time.Minute, 0)
	tc.mu.Lock()
	for i := 0; i < 100000; i++ {
		tc.set(strconv.Itoa(i), "bar", DefaultExpiration)
	}
	tc.mu.Unlock()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.DeleteExpired()
	}
}

func TestGetWithExpiration(t *testing.T) {
	tc := New[any](DefaultExpiration, 0)

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
	} else if a2 := (x).(int); a2+2 != 3 {
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
	} else if b2 := (x).(string); b2+"B" != "bB" {
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
	} else if c2 := (x).(float64); c2+1.2 != 4.7 {
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
	} else if d2 := (x).(int); d2+2 != 3 {
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
	} else if e2 := (x).(int); e2+2 != 3 {
		t.Error("e (which should be 1) plus 2 does not equal 3; value:", e2)
	}
	if expiration.UnixNano() != tc.items["e"].Expiration {
		t.Error("expiration for e is not the correct time")
	}
	if expiration.UnixNano() < time.Now().UnixNano() {
		t.Error("expiration for e is in the past")
	}
}
