package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/streadway/amqp"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
	"github.com/xinny/gateway/redis"
)

type guildChannel interface {
	GuildID() snowflake.ID
}

type extGuildChannel struct {
	discord.GuildChannel
}

func (e extGuildChannel) GuildID(id snowflake.ID) snowflake.ID {
	return id
}

type GuildCreateListener struct {
	client lib.GatewayClient
}

func (l GuildCreateListener) Run(ev gateway.EventData) {
	data := ev.(gateway.EventGuildCreate)
	guildId := data.ID.String()
	ctx := context.Background()
	oldData, err := l.client.Redis.Exists(ctx, fmt.Sprintf("%v:%v", common.GuildKey, data.ID.String())).Result()
	if err != nil {
		log.Fatalf("[%v] Couldn't perform EXISTS: %v", l.ListenerInfo().Event, err)
	}

	if oldData == 0 {
		body, _ := json.Marshal(data)
		err := l.client.BrokerChannel.Publish(l.client.BotApplication.ID.String(), string(l.ListenerInfo().Event), false, false, amqp.Publishing{
			Body: body,
		})
		if err != nil {
			log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
			return
		}
	}

	for _, v := range data.Members {
		v.GuildID = data.ID
		if *l.client.Config.State.User {
			userMap, err := redis.StructToMap(v.User)
			if err != nil {
				log.Fatalf("[%v] Couldn't convert struct to map: %v", l.ListenerInfo().Event, err)
			}
			stringifiedUser := redis.IterateMapAndStringify(userMap)
			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.UserKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, v.User.ID.String())).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				HSet(ctx, fmt.Sprintf("%v:%v", common.UserKey, v.User.ID.String()), stringifiedUser).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}

		if *l.client.Config.State.Member {
			if !*l.client.Config.State.User {
				v.User = discord.User{}
			}
			memberMap, err := redis.StructToMap(v)
			if err != nil {
				log.Fatalf("[%v] Couldn't convert struct to map: %v", l.ListenerInfo().Event, err)
			}
			stringifiedMember := redis.IterateMapAndStringify(memberMap)
			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.MemberKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, v.User.ID.String())).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				HSet(ctx, fmt.Sprintf("%v:%v:%v", common.MemberKey, guildId, v.User.ID.String()), stringifiedMember).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	if *l.client.Config.State.Role {
		for _, v := range data.Roles {
			v.GuildID = data.ID
			dataMap, err := redis.StructToMap(v)
			if err != nil {
				log.Fatalf("[%v] Couldn't convert struct to map: %v", l.ListenerInfo().Event, err)
			}
			stringified := redis.IterateMapAndStringify(dataMap)
			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.RoleKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, v.ID.String())).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				HSet(ctx, fmt.Sprintf("%v:%v:%v", common.RoleKey, guildId, v.ID.String()), stringified).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	if *l.client.Config.State.Voice {
		for _, v := range data.VoiceStates {
			v.GuildID = data.ID
			dataMap, err := redis.StructToMap(v)
			if err != nil {
				log.Fatalf("[%v] Couldn't convert struct to map: %v", l.ListenerInfo().Event, err)
			}
			stringified := redis.IterateMapAndStringify(dataMap)
			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.VoiceKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, v.UserID.String())).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				HSet(ctx, fmt.Sprintf("%v:%v:%v", common.VoiceKey, guildId, v.UserID.String()), stringified).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	if *l.client.Config.State.Channel {
		for _, v := range data.Channels {
			// TODO: Need to assign the guild id to this struct
			dataMap, err := redis.StructToMap(v)
			if err != nil {
				log.Fatalf("[%v] Couldn't convert struct to map: %v", l.ListenerInfo().Event, err)
			}
			stringified := redis.IterateMapAndStringify(dataMap)
			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.ChannelKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, v.ID().String())).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				HSet(ctx, fmt.Sprintf("%v:%v:%v", common.ChannelKey, guildId, v.ID().String()), stringified).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	if *l.client.Config.State.Emoji {
		for _, v := range data.Emojis {
			dataMap, err := redis.StructToMap(v)
			if err != nil {
				log.Fatalf("[%v] Couldn't convert struct to map: %v", l.ListenerInfo().Event, err)
			}
			stringified := redis.IterateMapAndStringify(dataMap)
			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.EmojiKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, v.ID.String())).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				HSet(ctx, fmt.Sprintf("%v:%v:%v", common.EmojiKey, guildId, v.ID.String()), stringified).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	if *l.client.Config.State.Presence {
		for _, v := range data.Presences {
			dataMap, err := redis.StructToMap(v)
			if err != nil {
				log.Fatalf("[%v] Couldn't convert struct to map: %v", l.ListenerInfo().Event, err)
			}
			stringified := redis.IterateMapAndStringify(dataMap)
			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.PresenceKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, v.PresenceUser.ID.String())).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				HSet(ctx, fmt.Sprintf("%v:%v:%v", common.PresenceKey, guildId, v.PresenceUser.ID.String()), stringified).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	data.Presences = []discord.Presence{}
	data.Channels = []discord.GuildChannel{}
	data.Emojis = []discord.Emoji{}
	data.VoiceStates = []discord.VoiceState{}
	data.Members = []discord.Member{}
	data.Roles = []discord.Role{}

	dataMap, err := redis.StructToMap(data)
	if err != nil {
		log.Fatalf("[%v] Couldn't convert struct to map: %v", l.ListenerInfo().Event, err)
	}
	stringified := redis.IterateMapAndStringify(dataMap)
	if _, err := l.client.Redis.
		HSet(ctx, fmt.Sprintf("%v:%v", common.GuildKey, data.ID.String()), stringified).
		Result(); err != nil {
		log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
	}
}

func (l GuildCreateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeGuildCreate,
	}
}

func RegisterGuildCreateListener(client *lib.GatewayClient) {
	common.RegisterListener(&GuildCreateListener{*client})
}
