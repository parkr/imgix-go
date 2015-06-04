# imgix-go

This is a Go implementation of an imgix url-building library outlined by
[imgix-blueprint](https://github.com/imgix/imgix-blueprint).

[Godoc](https://godoc.org/github.com/parkr/imgix-go)

[![Build Status](https://travis-ci.org/parkr/imgix-go.svg?branch=master)](https://travis-ci.org/parkr/imgix-go)

## Installation

It's a go package. Do this in your terminal:

```bash
go get github.com/parkr/imgix-go
```

## Usage

Something like this:

```go
package main

import (
    "fmt"
    "net/url"
    "github.com/parkr/imgix-go"
)

func main() {
    client := imgix.NewClient("mycompany.imgix.net")

    // Nothing fancy.
    fmt.Println(client.Path("/myImage.jpg"))

    // Throw some params in there!
    fmt.Println(client.PathWithParams("/myImage.jpg", url.Values{
        "w": []string{"400"},
        "h": []string{"400"},
    }))
}
```

That's it at a basic level. More fun features though!

### Sharding Hosts

This client supports sharding hosts, by CRC or just by Cycle.

**Cycle** is a simple round-robin algorithm. For each request, it picks the
host subsequent to the host for the previous request. Like this:

```go
client := Client{
  hosts: []string{"1.imgix.net", "2.imgix.net"},
  shardStrategy: imgix.ShardStrategyCycle,
}
client.Host("/myImage.jpg") // => uses 1.imgix.net
client.Host("/myImage.jpg") // => uses 2.imgix.net
client.Host("/myImage.jpg") // => uses 1.imgix.net... and so on.
```

**CRC** uses the CRC32 hashing algorithm paired with the input path to
determine the host. This allows you to ensure that an image request will
always hit the same host. It looks like this:

```go
client := Client{
  hosts: []string{"1.imgix.net", "2.imgix.net"},
  shardStrategy: imgix.ShardStrategyCRC,
}

// If you have the same path, you'll always get the same host.
client.Host("/myImage.jpg") // => uses 1.imgix.net
client.Host("/myImage.jpg") // => uses 1.imgix.net
client.Host("/myImage.jpg") // => uses 1.imgix.net... and so on.

// Now, a request for another image may find itself with a different host:
client.Host("/1/wedding.jpg") // => uses 2.imgix.net
client.Host("/1/wedding.jpg") // => uses 2.imgix.net
client.Host("/1/wedding.jpg") // => uses 2.imgix.net... and so on.
```

The default sharding is **Cycle**.
