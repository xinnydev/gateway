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

type MessageDeleteBulkListener struct {
	client lib.GatewayClient
}

func (l MessageDeleteBulkListener) Run(shardID int, ev gateway.EventData) {
	ctx := context.Background()
	data := ev.(gateway.EventMessageDeleteBulk)

	if *l.client.Config.State.Message {
		var messages []discord.Message
		keys, err := l.client.Redis.SMembers(ctx, fmt.Sprintf("%v%v", common.MessageKey, common.KeysSuffix)).Result()
		if err != nil {
			log.Fatalf("[%v] Couldn't perform SMEMBERS: %v", l.ListenerInfo().Event, err)
		}

		for _, v := range keys {
			var message discord.Message
			exists, err := l.client.Redis.HGetAllAndParse(fmt.Sprintf("%v:%v", common.MessageKey, v), &message)
			if err != nil {
				log.Fatalf("[%v] Couldn't perform HGetAllAndParse: %v", l.ListenerInfo().Event, err)
			}
			if exists {
				for _, m := range data.IDs {
					if m.String() == message.ID.String() {
						if _, err := l.client.Redis.
							SRem(ctx, fmt.Sprintf("%v%v", common.MessageKey, common.KeysSuffix), m).Result(); err != nil {
							log.Fatalf("[%v] Couldn't perform SREM: %v", l.ListenerInfo().Event, err)
						}

						if _, err := l.client.Redis.
							Unlink(ctx, fmt.Sprintf("%v:%v", common.MessageKey, v)).Result(); err != nil {
							log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
						}
						messages = append(messages, message)
					}
				}
			}
		}

		body, _ := json.Marshal(struct {
			Old []discord.Message `json:"old"`
			gateway.EventMessageDeleteBulk
		}{
			Old:                    messages,
			EventMessageDeleteBulk: data,
		})
		if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
			log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
			return
		}
	}

}

func (l MessageDeleteBulkListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeMessageDeleteBulk,
	}
}

func RegisterMessageDeleteBulkListener(client *lib.GatewayClient) {
	common.RegisterListener(&MessageDeleteBulkListener{*client})
}
