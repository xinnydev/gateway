package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type GuildRoleDeleteListener struct {
	client lib.GatewayClient
}

func (l GuildRoleDeleteListener) Run(ev gateway.EventData) {
	data := ev.(gateway.EventGuildRoleDelete)
	ctx := context.Background()
	guildId := data.GuildID.String()
	roleId := data.RoleID.String()

	if _, err := l.client.Redis.
		SRem(ctx, fmt.Sprintf("%v%v", common.RoleKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, roleId)).
		Result(); err != nil {
		log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
	}

	if _, err := l.client.Redis.
		Unlink(ctx, fmt.Sprintf("%v:%v:%v", common.RoleKey, guildId, roleId)).Result(); err != nil {
		log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
	}

	body, _ := json.Marshal(data)
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
