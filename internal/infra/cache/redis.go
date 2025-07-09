package cache

import "github.com/redis/go-redis/v9"

func NewRedisCache() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}
