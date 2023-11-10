package listener

import (
	"encoding/json"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
	"github.com/xinny/gateway/redis"
	"strconv"
)

type ReadyListener struct {
	client lib.GatewayClient
}

func (l ReadyListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventReady)
	if _, err := l.client.Redis.Hset(l.client.GenKey(common.BotUserKey), data.User); err != nil {
		log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
	}

	if _, err := l.client.Redis.Hset(l.client.GenKey(common.SessionKey, strconv.Itoa(data.Shard[0])), redis.SessionData{
		SessionID:      data.SessionID,
		ResumeURL:      data.ResumeGatewayURL,
		LastSequenceID: *l.client.ShardManager.Shard(shardID).LastSequenceReceived(),
	}); err != nil {
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

func (l ReadyListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeReady,
	}
}

func RegisterReadyListener(client *lib.GatewayClient) {
	common.RegisterListener(&ReadyListener{*client})
}
