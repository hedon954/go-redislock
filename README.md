# go-redislock
a distributed redis lock usable in a production environment implemented in Go

- [中文](./README_CN.md)



## How To Use

- Pull codes

  ```sh
  go get github.com/hedon954/go-redislock
  ```

  ```go
  import (
  	...
  	redislock "github.com/hedon954/go-redislock"
    ...
  )
  ```

- Init Client

  ```go
  yourRedisClient := ...
  c := redislock.NewCliengt(yourRedisClient)
  ```

- Prepare arguments according to your real business needs and call `Lock()`  or `TryLock()`

  ```go
  c.Lock(...)
  c.TryLock(...)
  ```

- If the concurrency is high, you can use the 'SingleFlightLock()' method to optimize the performance of distributed locks

  ```go
  c.SingleLightLock(...)
  ```



## Implementation Cores

- core: SETNX
- refresh expiration
- retry while outtime
- singleflight