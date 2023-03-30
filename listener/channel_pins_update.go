package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type ChannelPinsUpdateListener struct {
	client lib.GatewayClient
}

func (l ChannelPinsUpdateListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventChannelPinsUpdate)
	ctx := context.Background()
	if *l.client.Config.State.Channel {
		if data.GuildID != nil {
			channel, err := l.client.Redis.HGet(ctx, fmt.Sprintf("%v:%v:%v", common.ChannelKey, data.GuildID, data.ChannelID), "id").Result()
			if err != nil {
				log.Fatalf("[%v] Couldn't perform HGET: %v", l.ListenerInfo().Event, err)
			}
			if channel != "" {
				parsed, _ := json.Marshal(data.LastPinTimestamp)
				if _, err := l.client.Redis.HSet(ctx, fmt.Sprintf("%v:%v:%v", common.ChannelKey, data.GuildID, data.ChannelID), "last_pin_timestamp", string(parsed)).Result(); err != nil {
					log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
				}
			}
		} else {
			channel := l.client.Redis.HGet(context.Background(), fmt.Sprintf("%v:%v", common.ChannelKey, data.ChannelID), "id").String()
			if channel != "" {
				parsed, _ := json.Marshal(data.LastPinTimestamp)
				if _, err := l.client.Redis.HSet(ctx, fmt.Sprintf("%v:%v", common.ChannelKey, data.ChannelID), "last_pin_timestamp", string(parsed)).Result(); err != nil {
					log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
				}

			}
		}
	}
	body, _ := json.Marshal(data)
	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}
}

func (l ChannelPinsUpdateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeChannelPinsUpdate,
	}
}

func RegisterChannelPinsUpdateListener(client *lib.GatewayClient) {
	common.RegisterListener(&ChannelPinsUpdateListener{*client})
}
