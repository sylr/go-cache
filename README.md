# go-cache

[![Go Reference](https://pkg.go.dev/badge/sylr.dev/cache/v3.svg)](https://pkg.go.dev/sylr.dev/cache/v3)

## sylr.dev/cache fork disclaimer

This module is a fork of [github.com/patrickmn/go-cache/](https://github.com/patrickmn/go-cache/).

## Synopsys

go-cache is a generic in-memory key:value store/cache similar to memcached that is
suitable for applications running on a single machine. Its major advantage is
that, being essentially a thread-safe `map[string]interface{}` with expiration
times, it doesn't need to serialize or transmit its contents over the network.

Any object can be stored, for a given duration or forever, and the cache can be
safely used by multiple goroutines.

### Installation

```bash
go get sylr.dev/cache/v3
```

### Reference

The API reference can be found at [http://pkg.go.dev/sylr.dev/cache/v3](http://pkg.go.dev/sylr.dev/cache/v3).

### Examples

See [example_test.go](./example_test.go) for some usage examples.
