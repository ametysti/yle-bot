package db

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	rdb *redis.Client
)

// Connects the Redis instance, giving the Client and putting rdb to redis.go variables.
func Connect() *redis.Client {
	url := os.Getenv("REDIS_URI")

	if url == "" {
		panic("no REDIS_URI environment variable")
	}
	opts, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}

	rdb = redis.NewClient(opts)

	return rdb
}

func GetRecentID() string {
	data, _ := rdb.Get(context.Background(), "id").Result()

	return data

}

func SaveID(id string) bool {
	_, err := rdb.Set(context.Background(), "id", id, 0).Result()

	return err != nil
}
