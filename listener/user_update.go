package listener

import (
	"encoding/json"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type UserUpdateListener struct {
	client lib.GatewayClient
}

func (l UserUpdateListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventUserUpdate)
	if _, err := l.client.Redis.Hset(l.client.GenKey(common.BotUserKey), data.User); err != nil {
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

func (l UserUpdateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeUserUpdate,
	}
}

func RegisterUserUpdateListener(client *lib.GatewayClient) {
	common.RegisterListener(&UserUpdateListener{*client})
}
