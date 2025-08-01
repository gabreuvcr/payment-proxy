package infra

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func NewRedis() *redis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")

	return redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
}
