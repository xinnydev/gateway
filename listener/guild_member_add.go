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

type GuildMemberAddListener struct {
	client lib.GatewayClient
}

func (l GuildMemberAddListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventGuildMemberAdd)
	ctx := context.Background()
	guildId := data.GuildID.String()
	userId := data.User.ID.String()

	if *l.client.Config.State.User {
		if _, err := l.client.Redis.
			SAdd(ctx, fmt.Sprintf("%v%v", common.UserKey, common.KeysSuffix), userId).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
		}
		if _, err := l.client.Redis.
			Hset(fmt.Sprintf("%v:%v", common.UserKey, userId), data.User); err != nil {
			log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
		}
	}

	if *l.client.Config.State.Member {
		cloned := &data
		if !*l.client.Config.State.User {
			cloned.Member.User = discord.User{}
		}

		if _, err := l.client.Redis.
			SAdd(ctx, fmt.Sprintf("%v%v", common.MemberKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, userId)).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
		}
		if _, err := l.client.Redis.
			Hset(fmt.Sprintf("%v:%v:%v", common.MemberKey, guildId, userId), cloned); err != nil {
			log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
		}
	}

	body, _ := json.Marshal(data)
	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}
}

func (l GuildMemberAddListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeGuildMemberAdd,
	}
}

func RegisterGuildMemberAddListener(client *lib.GatewayClient) {
	common.RegisterListener(&GuildMemberAddListener{*client})
}
