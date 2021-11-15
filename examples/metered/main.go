package main

import (
	"fmt"
	"net/http"
	"time"

	"sylr.dev/cache/v3"
)

var (
	c = cache.NewNumericMeteredCacher[int64](time.Minute, 30*time.Second)
)

func main() {
	http.HandleFunc("/", incCache)
	http.ListenAndServe(fmt.Sprintf("%s:%d", "0", "8080"), nil)
}

func incCache(w http.ResponseWriter, r *http.Request) {
	c.
}
