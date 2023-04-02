package listener

import (
	"context"
	"fmt"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type RawListener struct {
	client lib.GatewayClient
}

func (l RawListener) Run(shardID int, ev gateway.EventData) {
	if _, err := l.client.Redis.HSet(context.Background(),
		fmt.Sprintf("%v:%v:%v", common.SessionKey, l.client.BotID, shardID), "last_seq",
		l.client.ShardManager.Shard(shardID).LastSequenceReceived()).Result(); err != nil {
		log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
	}
}

func (l RawListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeRaw,
	}
}

func RegisterRawListener(client *lib.GatewayClient) {
	common.RegisterListener(&RawListener{*client})
}
