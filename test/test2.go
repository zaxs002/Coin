package main

import (
	"github.com/garyburd/redigo/redis"
	"fmt"
)

func main() {
	c, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return
	}
	defer c.Close()


	for {
		c.Do("HMSET", "test", "name", "redis tutorial", "description", "redis basic commands", "symbol", 1343.4343)

		reply, err := c.Do("HGET", "HuoBi", "ethbtc")
		if err != nil {
			println(err)
			break
		}
		s, _ := redis.String(reply,err)
		println(s)
		i, _ := redis.Float64(c.Do("HGET", "HuoBi", "ethbtc"))
		fmt.Println(i)
	}
}
