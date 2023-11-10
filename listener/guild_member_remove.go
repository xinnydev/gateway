package listener

import (
	"context"
	"encoding/json"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type GuildMemberRemoveListener struct {
	client lib.GatewayClient
}

func (l GuildMemberRemoveListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventGuildMemberRemove)
	ctx := context.Background()

	body, _ := json.Marshal(&broker.PublishPayload{
		ShardID: shardID,
		Data:    data,
	})
	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}

	guildId := data.GuildID.String()
	userId := data.User.ID.String()

	if *l.client.Config.State.User {
		if _, err := l.client.Redis.
			SRem(ctx, l.client.GenKey(common.UserKey, common.KeysSuffix), userId).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform SREM: %v", l.ListenerInfo().Event, err)
		}
		if _, err := l.client.Redis.
			Unlink(ctx, l.client.GenKey(common.UserKey, userId)).Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
		}
	}

	if *l.client.Config.State.Member {
		if _, err := l.client.Redis.
			SRem(ctx, l.client.GenKey(common.MemberKey, common.KeysSuffix, guildId), userId).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform SREM: %v", l.ListenerInfo().Event, err)
		}
		if _, err := l.client.Redis.
			Unlink(ctx, l.client.GenKey(common.MemberKey, guildId, userId)).Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
		}
	}

	if *l.client.Config.State.Presence {
		if _, err := l.client.Redis.
			SRem(ctx, l.client.GenKey(common.PresenceKey, common.KeysSuffix, guildId), userId).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform SREM: %v", l.ListenerInfo().Event, err)
		}
		if _, err := l.client.Redis.
			Unlink(ctx, l.client.GenKey(common.PresenceKey, common.KeysSuffix, guildId, userId)).Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
		}
	}
}

func (l GuildMemberRemoveListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeGuildMemberRemove,
	}
}

func RegisterGuildMemberRemoveListener(client *lib.GatewayClient) {
	common.RegisterListener(&GuildMemberRemoveListener{*client})
}
