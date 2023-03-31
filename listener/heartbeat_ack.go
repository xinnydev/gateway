package listener

import (
	"context"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type HeartbeatAckListener struct {
	client lib.GatewayClient
}

func (l HeartbeatAckListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventHeartbeatAck)
	_, err := l.client.Redis.HSet(context.Background(), common.PingKey, shardID, data.NewHeartbeat.Sub(data.LastHeartbeat).Milliseconds()).Result()
	if err != nil {
		log.Fatalf("[%v] Couldn't perform Hget: %v", l.ListenerInfo().Event, err)
	}
}

func (l HeartbeatAckListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeHeartbeatAck,
	}
}

func RegisterHeartbeatAckListener(client *lib.GatewayClient) {
	common.RegisterListener(&HeartbeatAckListener{*client})
}
