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

type GuildRoleCreateListener struct {
	client lib.GatewayClient
}

func (l GuildRoleCreateListener) Run(ev gateway.EventData) {
	data := ev.(gateway.EventGuildRoleCreate)
	ctx := context.Background()
	guildId := data.GuildID.String()

	if *l.client.Config.State.Role {
		if _, err := l.client.Redis.
			SAdd(ctx, fmt.Sprintf("%v%v", common.RoleKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, data.Role.ID.String())).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
		}
		if _, err := l.client.Redis.
			Hset(fmt.Sprintf("%v:%v:%v", common.RoleKey, guildId, data.Role.ID.String()), data.Role); err != nil {
			log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
		}
	}

	body, _ := json.Marshal(data)
	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}
}

func (l GuildRoleCreateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeGuildRoleCreate,
	}
}

func RegisterGuildRoleCreateListener(client *lib.GatewayClient) {
	common.RegisterListener(&GuildRoleCreateListener{*client})
}
