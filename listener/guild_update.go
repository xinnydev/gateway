package listener

import (
	"encoding/json"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type GuildUpdateListener struct {
	client lib.GatewayClient
}

func (l GuildUpdateListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventGuildUpdate)
	guildId := data.Guild.ID.String()

	var old discord.Guild
	_, err := l.client.Redis.HGetAllAndParse(l.client.GenKey(common.GuildKey, guildId), &old)
	if err != nil {
		log.Fatalf("[%v] Couldn't perform HGetAllAndParse: %v", l.ListenerInfo().Event, err)
	}

	body, _ := json.Marshal(&broker.PublishPayload{
		ShardID: shardID,
		Data: struct {
			discord.Guild
			Old discord.Guild `json:"old"`
		}{
			Guild: data.Guild,
			Old:   old,
		},
	})
	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}

	if _, err := l.client.Redis.
		Hset(l.client.GenKey(common.GuildKey, data.ID.String()), data.Guild); err != nil {
		log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
	}
}

func (l GuildUpdateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeGuildUpdate,
	}
}

func RegisterGuildUpdateListener(client *lib.GatewayClient) {
	common.RegisterListener(&GuildUpdateListener{*client})
}
