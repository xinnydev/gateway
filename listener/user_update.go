package listener

import (
	"encoding/json"
	"fmt"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type UserUpdateListener struct {
	client lib.GatewayClient
}

func (l UserUpdateListener) Run(ev gateway.EventData) {
	data := ev.(gateway.EventUserUpdate)
	if _, err := l.client.Redis.Hset(fmt.Sprintf("%v:%v", common.BotUserKey, data.ID.String()), data.User); err != nil {
		log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
	}

	body, _ := json.Marshal(data)
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
