package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/xinny/gateway/common"
	"log"
	"strings"
)

type URLs []string

func (r *URLs) UnmarshalText(text []byte) error {
	*r = strings.Split(string(text), ",")
	return nil
}

type Config struct {
	Password *string `env:"REDIS_PASSWORD"`
	URLs     *URLs   `env:"REDIS_URL,required"`
	PoolSize *int    `env:"REDIS_POOL_SIZE,required"`
}

type Client struct {
	*redis.ClusterClient
}

func NewRedisClient(conf Config) *Client {
	opts := &redis.ClusterOptions{
		Addrs: *conf.URLs,
	}

	if conf.Password != nil {
		opts.Password = *conf.Password
	}

	client := redis.NewClusterClient(opts)

	return &Client{client}
}

func (c Client) Hset(key string, input interface{}) (int64, error) {
	mappedData, err := common.StructToMap(input)
	if err != nil {
		log.Fatalf("Couldn't convert struct to map: %v", err)
	}

	data := common.IterateMapAndStringify(mappedData)
	return c.HSet(context.Background(), key, data).Result()
}
