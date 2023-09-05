package cache

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"gpt-http/cfg"
)

var (
	redisCli *redis.Client
)

func init() {
	redisCli = redis.NewClient(&redis.Options{
		Addr:     cfg.Cfg.RedisAddr,
		Password: cfg.Cfg.RedisPasswd,
		//DB:       0,
	})
	_, err := redisCli.Ping().Result()
	if err != nil {
		panic("redis connect fail: " + err.Error())
	}
	fmt.Println("redis connected!")
}

func HSet(key, hKey string, v any) error {
	marshal, err := json.Marshal(v)
	if err != nil {
		fmt.Println(err)
		return err
	}
	redisCli.HSet(key, hKey, marshal)
	return nil
}

func HGet(key, hKey string, v any) error {
	get := redisCli.HGet(key, hKey)
	marshal, err := get.Bytes()
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = json.Unmarshal(marshal, v)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func HGetAll(key string) (map[string]string, error) {
	return redisCli.HGetAll(key).Val(), redisCli.HGetAll(key).Err()
}

func Keys() []string {
	return redisCli.Keys("*").Val()
}
