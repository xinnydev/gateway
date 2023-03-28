package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type GuildRoleUpdateListener struct {
	client lib.GatewayClient
}

func (l GuildRoleUpdateListener) Run(ev gateway.EventData) {
	data := ev.(gateway.EventGuildRoleUpdate)
	ctx := context.Background()
	guildId := data.GuildID.String()
	roleId := data.Role.ID.String()

	if *l.client.Config.State.Role {
		if _, err := l.client.Redis.
			SAdd(ctx, fmt.Sprintf("%v%v", common.RoleKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, roleId)).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
		}
		if _, err := l.client.Redis.
			Hset(fmt.Sprintf("%v:%v:%v", common.RoleKey, guildId, data.Role.ID.String()), data.Role); err != nil {
			log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
		}
	}

	var old *discord.Role
	exists, err := l.client.Redis.HGetAllAndParse(fmt.Sprintf("%v:%v:%v", common.RoleKey, guildId, roleId), &old)
	if err != nil {
		log.Fatalf("[%v] Couldn't perform HGetAllAndParse: %v", l.ListenerInfo().Event, err)
	}

	if exists {
		old = nil
	}

	body, _ := json.Marshal(struct {
		Old *discord.Role `json:"old"`
		discord.Role
	}{
		Role: data.Role,
		Old:  old,
	})
	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}
}

func (l GuildRoleUpdateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeGuildRoleUpdate,
	}
}

func RegisterGuildRoleUpdateListener(client *lib.GatewayClient) {
	common.RegisterListener(&GuildRoleUpdateListener{*client})
}
