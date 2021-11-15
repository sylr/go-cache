package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"sylr.dev/cache/v3"
)

var (
	c = cache.NewNumericMeteredCacher[int64](10*time.Second, 5*time.Second)
)

func main() {
	http.HandleFunc("/", incHitCache)
	http.HandleFunc("/get", getHitCache)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf("%s:%d", "0", 8080), nil)
}

func incHitCache(w http.ResponseWriter, r *http.Request) {
	if v, err := c.Increment("hit", 1); err != nil {
		fmt.Fprintf(w, "Error: %v", err)
		c.Set("hit", 1, 0)
	} else {
		fmt.Fprintf(w, "%d", v)
	}
}

func getHitCache(w http.ResponseWriter, r *http.Request) {
	hit, found := c.Get("hit")
	if found {
		fmt.Fprintf(w, "%d", hit)
	} else {
		fmt.Fprintf(w, "Not Found")
	}
}
