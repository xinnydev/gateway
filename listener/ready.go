package listener

import (
	"encoding/json"
	"fmt"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
	"github.com/xinny/gateway/redis"
)

type ReadyListener struct {
	client lib.GatewayClient
}

func (l ReadyListener) Run(ev gateway.EventData) {
	data := ev.(gateway.EventReady)
	clientId := data.User.ID.String()
	if _, err := l.client.Redis.Hset(fmt.Sprintf("%v:%v", common.BotUserKey, clientId), data.User); err != nil {
		log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
	}

	if _, err := l.client.Redis.Hset(fmt.Sprintf("%v:%v:%v", common.SessionKey, clientId, data.Shard[0]), redis.SessionData{
		SessionID: data.SessionID,
		ResumeURL: data.ResumeGatewayURL,
	}); err != nil {
		log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
	}
	body, _ := json.Marshal(data)
	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}
}

func (l ReadyListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeReady,
	}
}

func RegisterReadyListener(client *lib.GatewayClient) {
	common.RegisterListener(&ReadyListener{*client})
}
