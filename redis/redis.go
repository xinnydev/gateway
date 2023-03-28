package redis

import (
	"context"
	"encoding/json"
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

func (c Client) HGetAllAndParse(key string, output interface{}) error {
	ctx := context.Background()
	res, err := c.HGetAll(ctx, key).Result()
	if err != nil {
		log.Fatalf("Couldn't perform HGETALL: %v", err)
	}

	// Serialize types
	out := map[string]interface{}{}
	for k, v := range res {
		if v == "true" || v == "false" {
			out[k] = v == "true"
		} else if strings.HasSuffix(k, "id") && v == "null" {
			out[k] = 0
		} else if v == "null" || v == "undefined" {
			//parsed, _ := json.Marshal(v)
			out[k] = nil
		} else if (strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}")) ||
			(strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]")) {
			parsed, _ := json.Marshal(v)
			out[k] = parsed
		} else {
			out[k] = v
		}
	}

	// Need to convert stringified int
	//dst := reflect.TypeOf(output)
	//for j := 0; j < dst.NumField(); j++ {
	//	f := dst.Field(j)
	//	tag := f.Tag.Get("json")
	//	t := f.Type
	//	if v, ok := out[tag]; ok {
	//		if t.Kind() == reflect.Int {
	//			num, _ := strconv.Atoi(v.(string))
	//			out[tag] = num
	//		}
	//	}
	//}

	jsonData, _ := json.Marshal(out)
	log.Printf("%v\n", string(jsonData))
	err = json.Unmarshal(jsonData, output)
	if err != nil {
		return err
	}
	return nil
}
