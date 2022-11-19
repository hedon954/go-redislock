# go-redislock
一个用 Redis 实现的可在生产环境使用的分布式锁。

- [English](./README.md)



## 1. 如何使用

1. 拉取代码

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

2. 初始化 Client

   ```go
   yourRedisClient := ...
   c := redislock.NewCliengt(yourRedisClient)
   ```

3. 根据业务需求制定参数并调用 `Lock()`  或 `TryLock()` 方法

   ```go
   c.Lock(...)
   c.TryLock(...)
   ```

4. 如果并发量很高，可以使用 `SingleFlightLock()` 方法来优化分布式锁的性能

   ```go
   c.SingleLightLock(...)
   ```



## 2. 实现要点

1. 核心：SETNX
2. 续约
3. 超时重试
4. singleflight