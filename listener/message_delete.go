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

type MessageDeleteListener struct {
	client lib.GatewayClient
}

func (l MessageDeleteListener) Run(shardID int, ev gateway.EventData) {
	ctx := context.Background()
	data := ev.(gateway.EventMessageDelete)
	guildId := data.GuildID.String()
	msgId := data.ID.String()

	var old discord.Message
	_, err := l.client.Redis.HGetAllAndParse(fmt.Sprintf("%v:%v:%v", common.MessageKey, guildId, msgId), &old)
	if err != nil {
		log.Fatalf("[%v] Couldn't perform HGetAllAndParse: %v", l.ListenerInfo().Event, err)
	}

	body, _ := json.Marshal(struct {
		Old discord.Message `json:"old"`
		gateway.EventMessageDelete
	}{
		Old:                old,
		EventMessageDelete: data,
	})

	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}

	if *l.client.Config.State.Message {
		if _, err := l.client.Redis.
			SRem(ctx, fmt.Sprintf("%v%v", common.MessageKey, common.KeysSuffix), fmt.Sprintf("%v:%v", guildId, msgId)).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform SREM: %v", l.ListenerInfo().Event, err)
		}

		if _, err := l.client.Redis.
			Unlink(ctx, fmt.Sprintf("%v:%v:%v", common.MessageKey, guildId, msgId)).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
		}
	}
}

func (l MessageDeleteListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeMessageDelete,
	}
}

func RegisterMessageDeleteListener(client *lib.GatewayClient) {
	common.RegisterListener(&MessageDeleteListener{*client})
}
