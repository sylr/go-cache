package cache_test

import (
	"fmt"
	"time"

	"sylr.dev/cache/v3"
)

// -- AnyCache -----------------------------------------------------------------

type MyStruct struct {
	Name string
}

func ExampleAnyCache_any() {
	// Create a generic cache with a default expiration time of 5 minutes
	// All items are cached as interface{} so they need to be cast back to their
	// original type when retrieved.
	c := cache.NewAnyCacher[any](5*time.Minute, 10*time.Minute)

	myStruct := &MyStruct{"MySuperStruct"}

	c.Set("MySuperStruct", myStruct, 0)

	myRawCachedStruct, found := c.Get("MySuperStruct")

	if found {
		// Casting the retrieved object back to its original type
		myCachedStruct := myRawCachedStruct.(*MyStruct)
		fmt.Printf("%s", myCachedStruct.Name)
	} else {
		fmt.Printf("Error: MySuperStruct not found in cache")
	}

	// Output:
	// MySuperStruct
}

func ExampleAnyCacher_string() {
	// Create a string cache with a default expiration time of 5 minutes.
	c := cache.NewAnyCacher[string](5*time.Minute, 10*time.Minute)

	c.Set("string", "MySuperString", 0)

	mySuperString, found := c.Get("string")

	if found {
		fmt.Printf("%s", mySuperString)
	} else {
		fmt.Printf("Error: MySuperStruct not found in cache")
	}

	// Output:
	// MySuperString
}

func ExampleAnyCacher_customStruct() {
	// Create a cache with a default expiration time of 5 minutes, and which
	c := cache.NewAnyCacher[*MyStruct](5*time.Minute, 10*time.Minute)

	myStruct := &MyStruct{"MySuperStruct"}

	c.Set("MySuperStruct", myStruct, 0)

	myCachedStruct, found := c.Get("MySuperStruct")

	if found {
		fmt.Printf("%s", myCachedStruct.Name)
	} else {
		fmt.Printf("Error: MySuperStruct not found in cache")
	}

	// Output:
	// MySuperStruct
}

// -- NumericCache -------------------------------------------------------------

func ExampleNumericCacher_int8() {
	// Create a float64 cache with a default expiration time of 5 minutes.
	c := cache.NewNumericCacher[int8](5*time.Minute, 10*time.Minute)
	key := "universeAnswer"

	c.Set(key, 42, 0)
	c.Increment(key, 1)
	c.Decrement(key, 1)

	universeAnswer, found := c.Get("universeAnswer")

	if found {
		fmt.Printf("%d", universeAnswer)
	} else {
		fmt.Printf("Error: universeAnswer not found in cache")
	}

	// Output:
	// 42
}

func ExampleNumericCache_float64() {
	// Create a float64 cache with a default expiration time of 5 minutes.
	c := cache.NewNumericCacher[float64](5*time.Minute, 10*time.Minute)
	key := "universeAnswer"

	c.Set(key, 42.0, 0)
	c.Increment(key, 1.0)
	c.Decrement(key, 1.0)

	universeAnswer, found := c.Get(key)

	if found {
		fmt.Printf("%.1f", universeAnswer)
	} else {
		fmt.Printf("Error: %s not found in cache", key)
	}

	// Output:
	// 42.0
}
