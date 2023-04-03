package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type GuildCreateListener struct {
	client lib.GatewayClient
}

func (l GuildCreateListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventGuildCreate)
	guildId := data.ID.String()
	ctx := context.Background()
	oldData, err := l.client.Redis.Exists(ctx, fmt.Sprintf("%v:%v", common.GuildKey, data.ID.String())).Result()
	if err != nil {
		log.Fatalf("[%v] Couldn't perform EXISTS: %v", l.ListenerInfo().Event, err)
	}

	if oldData == 0 {
		body, _ := json.Marshal(&broker.PublishPayload{
			ShardID: shardID,
			Data:    data,
		})
		if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
			log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
			return
		}
	}

	for _, v := range data.Members {
		v.GuildID = data.ID
		if *l.client.Config.State.User {
			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.UserKey, common.KeysSuffix), v.User.ID.String()).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				Hset(fmt.Sprintf("%v:%v", common.UserKey, v.User.ID.String()), v); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}

		if *l.client.Config.State.Member {
			memberId := v.User.ID
			cloned := &v
			if !*l.client.Config.State.User {
				cloned.User = discord.User{}
			}

			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.MemberKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, memberId)).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				Hset(fmt.Sprintf("%v:%v:%v", common.MemberKey, guildId, memberId), cloned); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	if *l.client.Config.State.Role {
		for _, v := range data.Roles {
			v.GuildID = data.ID
			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.RoleKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, v.ID.String())).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				Hset(fmt.Sprintf("%v:%v:%v", common.RoleKey, guildId, v.ID.String()), v); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	if *l.client.Config.State.Voice {
		for _, v := range data.VoiceStates {
			v.GuildID = data.ID
			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.VoiceKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, v.UserID.String())).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				Hset(fmt.Sprintf("%v:%v:%v", common.VoiceKey, guildId, v.UserID.String()), v); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	if *l.client.Config.State.Channel {
		for _, v := range data.Channels {
			dataMap, err := common.StructToMap(v)
			if err != nil {
				log.Fatalf("[%v] Couldn't convert struct to map: %v", l.ListenerInfo().Event, err)
			}
			// Need to exclude and append guild_id manually
			// since the struct that returning guild id is interface
			stringified := common.IterateMapAndStringify(dataMap, "guild_id")
			stringified = append(stringified, "guild_id", guildId)
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
			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.EmojiKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, v.ID.String())).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				Hset(fmt.Sprintf("%v:%v:%v", common.EmojiKey, guildId, v.ID.String()), v); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	if *l.client.Config.State.Presence {
		for _, v := range data.Presences {
			if _, err := l.client.Redis.
				SAdd(ctx, fmt.Sprintf("%v%v", common.PresenceKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, v.PresenceUser.ID.String())).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.
				Hset(fmt.Sprintf("%v:%v:%v", common.PresenceKey, guildId, v.PresenceUser.ID.String()), v); err != nil {
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

	if _, err := l.client.Redis.
		Hset(fmt.Sprintf("%v:%v", common.GuildKey, data.ID.String()), data); err != nil {
		log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
	}

	if _, err := l.client.Redis.
		SAdd(ctx, fmt.Sprintf("%v%v", common.GuildKey, common.KeysSuffix), guildId).Result(); err != nil {
		log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
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
