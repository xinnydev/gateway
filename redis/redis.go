package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/xinny/gateway/common"
	"log"
	"reflect"
	"strconv"
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

// HGetAllAndParse FIXME: This function may be inefficient but idfc, as long it works :handshake:
// but, seriously, this need a refactor if someone found the better way
func (c Client) HGetAllAndParse(key string, output interface{}) (exists bool, err error) {
	ctx := context.Background()
	res, err := c.HGetAll(ctx, key).Result()
	if err != nil {
		log.Fatalf("Couldn't perform HGETALL: %v", err)
	}

	if len(res) == 0 {
		return false, nil
	}

	// Serialize types
	out := map[string]interface{}{}
firstLoop:
	for k, v := range res {
		structElem := reflect.TypeOf(output).Elem()
		for s := 0; s < structElem.NumField(); s++ {
			field := structElem.Field(s)
			tag := field.Tag.Get("json")
			fType := field.Type.Kind()
			if tag != k {
				continue
			}

			// If the struct field typed int, we have to parse the string manually
			if fType == reflect.Int {
				num, _ := strconv.Atoi(v)
				out[k] = num
				continue firstLoop
			}
		}

		// Parse boolean
		if v == "true" || v == "false" {
			out[k] = v == "true"
		} else if (strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}")) ||
			(strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]")) ||
			v == "null" {
			// Parse stringified array or json such as: "[]" or "{}"
			// and the null value
			var result struct {
				Res interface{} `json:"res,omitempty"`
			}
			_ = json.Unmarshal([]byte(fmt.Sprintf("{ \"res\": %v }", v)), &result)
			out[k] = result.Res
		} else {
			// We assume the rest is an ordinary string, no need to parse
			out[k] = v
		}
	}

	jsonData, _ := json.Marshal(out)
	err = json.Unmarshal(jsonData, output)
	if err != nil {
		return true, err
	}
	return true, nil
}
