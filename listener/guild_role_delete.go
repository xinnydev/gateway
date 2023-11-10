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

type GuildRoleDeleteListener struct {
	client lib.GatewayClient
}

func (l GuildRoleDeleteListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventGuildRoleDelete)
	ctx := context.Background()
	guildId := data.GuildID.String()
	roleId := data.RoleID.String()

	if _, err := l.client.Redis.
		SRem(ctx, l.client.GenKey(common.RoleKey, common.KeysSuffix, guildId), roleId).
		Result(); err != nil {
		log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
	}

	if _, err := l.client.Redis.
		Unlink(ctx, l.client.GenKey(common.RoleKey, guildId, roleId)).Result(); err != nil {
		log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
	}

	body, _ := json.Marshal(&broker.PublishPayload{
		ShardID: shardID,
		Data:    data,
	})
	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}
}

func (l GuildRoleDeleteListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeGuildRoleDelete,
	}
}

func RegisterGuildRoleDeleteListener(client *lib.GatewayClient) {
	common.RegisterListener(&GuildRoleDeleteListener{*client})
}
