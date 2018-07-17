package cache

import (
	"github.com/go-redis/redis"
	"sync"
	"fmt"
)

var RedisClient *redis.Client
var once sync.Once

var redisServerUrl = "127.0.0.1:6379"
//var redisServerUrl = "202.182.126.103:6379"
func GetInstance() *redis.Client {
	once.Do(func() {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     redisServerUrl,
			Password: "",
			DB:       0,
		})
		pong, err := RedisClient.Ping().Result()
		if err != nil {
			fmt.Println(pong, err)
		}
	})
	return RedisClient
}
