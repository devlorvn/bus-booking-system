package redis

import (
	"booking-system/configs"
	"context"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient(cfg *configs.Redis) *goredis.Client {
	addr := cfg.Host + ":" + cfg.Port
	client := goredis.NewClient(&goredis.Options{
		Addr: addr,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	return client
}
