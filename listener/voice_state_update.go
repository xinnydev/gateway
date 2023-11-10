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
)

type VoiceStateUpdateListener struct {
	client lib.GatewayClient
}

func (l VoiceStateUpdateListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventVoiceStateUpdate)
	ctx := context.Background()
	guildId := data.GuildID.String()
	userId := data.UserID.String()

	var old discord.VoiceState
	exists, err := l.client.Redis.HGetAllAndParse(l.client.GenKey(common.VoiceKey, guildId, userId), &old)
	if err != nil {
		log.Fatalf("[%v] Couldn't perform HGetAllAndParse: %v", l.ListenerInfo().Event, err)
	}

	var oldPayload *discord.VoiceState
	if exists {
		oldPayload = &old
	} else {
		oldPayload = nil
	}

	payload, _ := json.Marshal(&broker.PublishPayload{
		ShardID: shardID,
		Data: struct {
			Old *discord.VoiceState `json:"old,omitempty"`
			discord.VoiceState
		}{
			Old:        oldPayload,
			VoiceState: data.VoiceState,
		},
	})

	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), payload); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}

	if data.ChannelID == nil {
		if _, err := l.client.Redis.
			SRem(ctx, l.client.GenKey(common.VoiceKey, common.KeysSuffix, guildId), userId).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform SREM: %v", l.ListenerInfo().Event, err)
		}
		if _, err := l.client.Redis.
			Unlink(ctx, l.client.GenKey(common.VoiceKey, guildId, userId)).Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
		}
	} else {
		if _, err := l.client.Redis.
			SAdd(ctx, l.client.GenKey(common.VoiceKey, common.KeysSuffix, guildId), userId).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
		}
		if _, err := l.client.Redis.
			Hset(l.client.GenKey(common.VoiceKey, guildId, userId), data.VoiceState); err != nil {
			log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
		}
	}
}

func (l VoiceStateUpdateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeVoiceStateUpdate,
	}
}

func RegisterVoiceStateUpdateListener(client *lib.GatewayClient) {
	common.RegisterListener(&VoiceStateUpdateListener{*client})
}
