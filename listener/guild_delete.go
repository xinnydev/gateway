package listener

import (
	"context"
	"encoding/json"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
	"strings"
)

type GuildDeleteListener struct {
	client lib.GatewayClient
}

func (l GuildDeleteListener) clearGuildCache(guildId, redisKey string, data []string) {
	ctx := context.Background()
	for _, v := range data {
		if !strings.HasPrefix(v, guildId) {
			continue
		}
		l.client.Redis.Unlink(ctx, l.client.GenKey(redisKey, guildId))
		l.client.Redis.
			SRem(ctx, l.client.GenKey(redisKey, common.KeysSuffix, guildId), v)
	}
}
func (l GuildDeleteListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventGuildDelete)
	ctx := context.Background()
	guildId := data.Guild.ID.String()
	roles, err := l.client.Redis.SMembers(ctx, l.client.GenKey(common.RoleKey, common.KeysSuffix, guildId)).
		Result()
	if err != nil {
		log.Fatalf("[%v] Couldn't perform SMEMBERS: %v", l.ListenerInfo().Event, err)
	}

	members, err := l.client.Redis.SMembers(ctx, l.client.GenKey(common.MemberKey, common.KeysSuffix, guildId)).
		Result()
	if err != nil {
		log.Fatalf("[%v] Couldn't perform SMEMBERS: %v", l.ListenerInfo().Event, err)
	}

	users, err := l.client.Redis.SMembers(ctx, l.client.GenKey(common.UserKey, common.KeysSuffix, guildId)).
		Result()
	if err != nil {
		log.Fatalf("[%v] Couldn't perform SMEMBERS: %v", l.ListenerInfo().Event, err)
	}

	channels, err := l.client.Redis.SMembers(ctx, l.client.GenKey(common.ChannelKey, common.KeysSuffix, guildId)).
		Result()
	if err != nil {
		log.Fatalf("[%v] Couldn't perform SMEMBERS: %v", l.ListenerInfo().Event, err)
	}

	voices, err := l.client.Redis.SMembers(ctx, l.client.GenKey(common.VoiceKey, common.KeysSuffix, guildId)).
		Result()
	if err != nil {
		log.Fatalf("[%v] Couldn't perform SMEMBERS: %v", l.ListenerInfo().Event, err)
	}

	presences, err := l.client.Redis.SMembers(ctx, l.client.GenKey(common.PresenceKey, common.KeysSuffix, guildId)).
		Result()
	if err != nil {
		log.Fatalf("[%v] Couldn't perform SMEMBERS: %v", l.ListenerInfo().Event, err)
	}

	l.clearGuildCache(guildId, common.RoleKey, roles)
	l.clearGuildCache(guildId, common.MemberKey, members)
	l.clearGuildCache(guildId, common.UserKey, users)
	l.clearGuildCache(guildId, common.PresenceKey, presences)
	l.clearGuildCache(guildId, common.VoiceKey, voices)
	l.clearGuildCache(guildId, common.ChannelKey, channels)

	if !data.Unavailable {
		var oldGuild discord.Guild
		exists, err := l.client.Redis.HGetAllAndParse(l.client.GenKey(common.GuildKey, guildId), &oldGuild)
		if err != nil {
			log.Fatalf("[%v] Couldn't perform HGetAllAndParse: %v", l.ListenerInfo().Event, err)
		}

		if !exists {
			log.Fatalf("[%v] Guild cache expected to present", l.ListenerInfo().Event, err)
		}

		payload, _ := json.Marshal(&broker.PublishPayload{
			ShardID: shardID,
			Data: struct {
				Old discord.Guild `json:"old"`
				gateway.EventGuildDelete
			}{
				EventGuildDelete: data,
				Old:              oldGuild,
			},
		})

		if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), payload); err != nil {
			log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
			return
		}
	}

	if _, err := l.client.Redis.
		SRem(ctx, l.client.GenKey(common.GuildKey, common.KeysSuffix), guildId).Result(); err != nil {
		log.Fatalf("[%v] Couldn't perform SREM: %v", l.ListenerInfo().Event, err)
	}

	if _, err := l.client.Redis.
		Unlink(ctx, l.client.GenKey(common.GuildKey, common.GuildKey)).Result(); err != nil {
		log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
	}
}

func (l GuildDeleteListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeGuildDelete,
	}
}

func RegisterGuildDeleteListener(client *lib.GatewayClient) {
	common.RegisterListener(&GuildDeleteListener{*client})
}
