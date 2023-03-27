package redis

import (
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slices"
	"reflect"
	"strings"
)

type RedisURLs []string

func (r *RedisURLs) UnmarshalText(text []byte) error {
	*r = strings.Split(string(text), ",")
	return nil
}

type RedisConfig struct {
	Password *string    `env:"REDIS_PASSWORD"`
	URLs     *RedisURLs `env:"REDIS_URL,required"`
	PoolSize *int       `env:"REDIS_POOL_SIZE,required"`
}

type RedisClient struct {
	*redis.ClusterClient
}

func NewRedisClient(conf RedisConfig) *RedisClient {
	opts := &redis.ClusterOptions{
		Addrs: *conf.URLs,
	}

	if conf.Password != nil {
		opts.Password = *conf.Password
	}

	client := redis.NewClusterClient(opts)

	return &RedisClient{client}
}

func StructToMap(obj interface{}) (newMap map[string]interface{}, err error) {
	data, err := json.Marshal(obj) // Convert to a json string

	if err != nil {
		return
	}

	err = json.Unmarshal(data, &newMap) // Convert to a map
	return
}

func IterateMapAndStringify(in map[string]interface{}, exclude ...string) []string {
	var out []string
	for k, val := range in {
		if slices.Contains(exclude, k) {
			continue
		}

		v := reflect.ValueOf(val)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() == 0 {
			out = append(out, k, "null")
			continue
		}

		if v.Kind() == reflect.String {
			out = append(out, k, v.Interface().(string))
			continue
		}

		stringified, _ := json.Marshal(val)
		out = append(out, k, string(stringified))
	}
	return out
}
