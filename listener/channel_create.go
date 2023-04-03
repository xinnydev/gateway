package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type ChannelCreateListener struct {
	client lib.GatewayClient
}

func (l ChannelCreateListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventChannelCreate)
	ctx := context.Background()
	if *l.client.Config.State.Channel {
		if data.Type() == discord.ChannelTypeDM {
			if _, err := l.client.Redis.SAdd(ctx, fmt.Sprintf("%v%v", common.ChannelKey, common.KeysSuffix), data.ID()).Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.Hset(fmt.Sprintf("%v:%v", common.ChannelKey, data.ID()), data); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		} else {
			if _, err := l.client.Redis.SAdd(
				context.Background(),
				fmt.Sprintf("%v%v", common.ChannelKey, common.KeysSuffix),
				fmt.Sprintf("%v:%v", data.GuildID(), data.ID())).Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.Hset(fmt.Sprintf("%v:%v:%v", common.ChannelKey, data.GuildID(), data.ID()), data); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
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

func (l ChannelCreateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeChannelCreate,
	}
}

func RegisterChannelCreateListener(client *lib.GatewayClient) {
	common.RegisterListener(&ChannelCreateListener{*client})
}
