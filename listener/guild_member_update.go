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

type GuildMemberUpdateListener struct {
	client lib.GatewayClient
}

func (l GuildMemberUpdateListener) Run(ev gateway.EventData) {
	data := ev.(gateway.EventGuildMemberUpdate)
	ctx := context.Background()
	guildId := data.GuildID.String()
	userId := data.User.ID.String()

	var old discord.Member
	_, err := l.client.Redis.HGetAllAndParse(fmt.Sprintf("%v:%v:%v", common.MemberKey, guildId, userId), &old)
	if err != nil {
		log.Fatalf("[%v] Couldn't perform HGetAllAndParse: %v", l.ListenerInfo().Event, err)
	}

	body, _ := json.Marshal(struct {
		Old discord.Member `json:"old"`
		discord.Member
	}{
		Member: data.Member,
		Old:    old,
	})
	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}

	if *l.client.Config.State.User || data.User.ID.String() == l.client.BotID {
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

	if *l.client.Config.State.Member || data.User.ID.String() == l.client.BotID {
		if _, err := l.client.Redis.
			SAdd(ctx, fmt.Sprintf("%v%v", common.MemberKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, userId)).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
		}

		toSet := &data.Member
		if !*l.client.Config.State.User {
			toSet.User = discord.User{}
		}

		if _, err := l.client.Redis.
			Hset(fmt.Sprintf("%v:%v:%v", common.MemberKey, guildId, userId), toSet); err != nil {
			log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
		}
	}
}

func (l GuildMemberUpdateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeGuildMemberUpdate,
	}
}

func RegisterGuildMemberUpdateListener(client *lib.GatewayClient) {
	common.RegisterListener(&GuildMemberUpdateListener{*client})
}
