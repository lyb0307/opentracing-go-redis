# opentracing-go-redis

[OpenTracing](http://opentracing.io/) instrumentation for [go-redis](https://github.com/go-redis/redis).

This instrumentation bases on internal tracing of [go-redis](https://github.com/go-redis/redis/blob/master/redisext/otel.go).

## Install

```
go get -u github.com/lyb0307/opentracing-go-redis
```

## Usage

Example:

```go

package main

import (
"github.com/go-redis/redis/v8"
"github.com/lyb0307/otgoredis"
)

func main() {
    tracer := ... // init your tracer

    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB: 0,
    })
    
    rdb.AddHook(otgoredis.NewHookWithTracer(tracer))
}
```

## License

[MIT](LICENSE)