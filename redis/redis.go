package redis

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/disgoorg/log"
	"github.com/redis/go-redis/v9"
	"github.com/xinny/gateway/common"
	"reflect"
	"strconv"
	"strings"
)

type URLs []string
type SessionData struct {
	SessionID      string `json:"session_id"`
	ResumeURL      string `json:"resume_url"`
	LastSequenceID int    `json:"last_sequence_id"`
}

func (r *URLs) UnmarshalText(text []byte) error {
	*r = strings.Split(string(text), ",")
	return nil
}

type Config struct {
	Password *string `env:"REDIS_PASSWORD" envDefault:""`
	Username *string `env:"REDIS_USERNAME" envDefault:""`
	URLs     *URLs   `env:"REDIS_URL,required"`
	PoolSize *int    `env:"REDIS_POOL_SIZE"`
	Db       *int    `env:"REDIS_DB" envDefault:"0"`
}

type Client struct {
	config Config
	redis.UniversalClient
}

func NewRedisClient(conf Config) *Client {
	opts := &redis.UniversalOptions{
		Addrs:          *conf.URLs,
		RouteByLatency: true,
		DB:             *conf.Db,
	}

	if conf.Username != nil {
		opts.Username = *conf.Username
	}

	if conf.Password != nil {
		opts.Password = *conf.Password
	}

	if conf.PoolSize != nil {
		opts.PoolSize = *conf.PoolSize
	}

	client := redis.NewUniversalClient(opts)
	return &Client{UniversalClient: client, config: conf}
}

func (c Client) Hset(key string, input interface{}) (int64, error) {
	mappedData, err := common.StructToMap(input)
	if err != nil {
		log.Fatalf("Couldn't convert struct to map: %v", err)
	}

	data := common.IterateMapAndStringify(mappedData)
	if len(data) == 0 {
		log.Fatalf("[redis] unable to stringify the data. final map length: %v", len(data))
	}

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

func (c Client) ScanKeys(pattern string) (result []string, err error) {
	ctx := context.Background()
	if len(*c.config.URLs) > 1 {
		cluster := (c.UniversalClient).(*redis.ClusterClient)
		err = cluster.ForEachShard(ctx, func(ctx context.Context, node *redis.Client) error {
			iter := node.Scan(ctx, 0, pattern, 0).Iterator()
			for iter.Next(ctx) {
				result = append(result, iter.Val())
			}
			return nil
		})
		return
	}

	iter := c.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		result = append(result, iter.Val())
	}
	return
}
