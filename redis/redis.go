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
	SessionID string `json:"session_id"`
	ResumeURL string `json:"resume_url"`
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
		Addrs:         *conf.URLs,
		RouteRandomly: true,
		DB:            *conf.Db,
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
	return c.HSet(context.Background(), key, data).Result()
}

func (c Client) ClearCache() {
	patterns := []string{
		common.RoleKey,
		common.MemberKey,
		common.UserKey,
		common.MessageKey,
		common.VoiceKey,
		common.GuildKey,
		common.PresenceKey,
		common.EmojiKey,
		common.ChannelKey,
		common.SessionKey}

	for _, v := range patterns {
		// Clear Hash
		keys, err := c.SMembers(context.Background(), fmt.Sprintf("%v%v", v, common.KeysSuffix)).Result()
		if err != nil {
			log.Fatalf("[clearCache] unable to scan keys: %v", err)
		}
		if len(keys) < 1 {
			log.Infof("[clearCache] %v:* is empty", v)
			continue
		}

		for i := 0; i < len(keys); i++ {
			keys[i] = fmt.Sprintf("%v:%v", v, keys[i])
		}

		res, err := c.Unlink(context.Background(), keys...).Result()
		if err != nil {
			log.Fatalf("[clearCache] unable to unlink keys: %v", err)
		}
		log.Infof("[clearCache] unlinked %v %v:*", res, v)

		// Clear Set
		res, err = c.Unlink(context.Background(), fmt.Sprintf("%v%v", v, common.KeysSuffix)).Result()
		if err != nil {
			log.Fatalf("[clearCache] unable to unlink keys: %v", err)
		}
		if res != 0 {
			log.Infof("[clearCache] unlinked %v %v%v", res, v, common.KeysSuffix)
		}
	}
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
