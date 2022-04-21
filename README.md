# grocery [![Build Status](https://drone.dv.nyt.net/api/badges/nytimes/grocery/status.svg)](https://drone.dv.nyt.net/nytimes/grocery) [![Go Reference](https://pkg.go.dev/badge/github.com/nytimes/grocery.svg)](https://pkg.go.dev/github.com/nytimes/grocery)

<p align="center">
  <img src="docs/gopher.png" width="128" />
</p>

**Grocery** is a framework for simple object storage with Redis in Go.

This package adds helpful primitives on top of the popular [go-redis](https://github.com/go-redis/redis) framework to store and load structs with ease. Within these structs, Grocery provides built-in support for storing lists, sets, pointers to other structs, and other custom datatypes.

## Install

```bash
$ go get github.com/nytimes/grocery
```

## Test

To run tests, you must first have a Redis server running on `localhost:6379`. If you have Docker installed, you may start one with `docker run -dp 6379:6379 redis`.

```bash
$ go test
```

## Example

```go
package main

import (
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/nytimes/grocery"
)

// Supplier specifies how suppliers should be stored in Redis.
type Supplier struct {
	// Provides createdAt, updatedAt, and ID
	grocery.Base

	// All primitive types are supported. Grocery uses the field's name as the
	// key, with the first letter lowercased. Without grocery, this value could
	// later be retrieved with "HGET supplier:<id> name"
	Name string
}

// Fruit specifies how fruits should be stored in Redis.
type Fruit struct {
	grocery.Base
	Name string

	// If another key is desired, it can be specified with a field tag.
	// This value can be retrieved with "HGET fruit:<id> cost"
	Price float64 `grocery:"cost"`

	// Basic structures such as maps and lists are supported out of the
	// box as well. Maps are stored as their own key, so this value can be
	// retrieved with "HGETALL fruit:<id>:metadata"
	Metadata *grocery.Map

	// Pointers to other structs are supported. In Redis, the ID to the
	// struct is stored as a string. When this fruit is loaded with
	// grocery.Load, it will load the supplier struct as well.
	Supplier *Supplier
}

func main() {
	grocery.Init(&redis.Options{
		Addr: "localhost:6379",
	})

	supplier := &Supplier{
		Name: "R&D Fruit Co.",
	}

	id, _ := grocery.Store(supplier)
	fmt.Printf("Stored at supplier:%s\n", id)

	kv := sync.Map{}
	kv.Store("weight", 4.6)

	fruit := &Fruit{
		Name: "mango",
		Metadata: &grocery.Map{Map: kv},
		Supplier: supplier,
	}

	id, _ := grocery.Store(fruit)
	fmt.Printf("Stored at fruit:%s\n", id)
}
```

## Contributing

Refer to [CONTRIBUTING.md](./CONTRIBUTING.md) for general contribution instructions.

## License

grocery is available under the Apache 2.0 license. See the LICENSE file for details.

---

> This repository is maintained by the Research & Development team at The New York Times and is provided as-is for your own use. For more information about R&D at the Times visit [rd.nytimes.com](https://rd.nytimes.com)
