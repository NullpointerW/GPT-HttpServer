package cache

import (
	"fmt"
	"github.com/go-redis/redis"
	"gpt3.5/cfg"
)

var (
	Redis *redis.Client
)

func init() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     cfg.Cfg.RedisAddr,
		Password: cfg.Cfg.RedisPasswd,
		//DB:       0,
	})
	_, err := Redis.Ping().Result()
	if err != nil {
		panic("redis connect fail: " + err.Error())
	}
	fmt.Println("redis connected!")
}
