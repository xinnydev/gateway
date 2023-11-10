package listener

import (
	"context"
	"encoding/json"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type ChannelPinsUpdateListener struct {
	client lib.GatewayClient
}

func (l ChannelPinsUpdateListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventChannelPinsUpdate)
	ctx := context.Background()
	channelId := data.ChannelID.String()
	if *l.client.Config.State.Channel {
		if data.GuildID != nil {
			guildId := data.GuildID.String()
			channel, err := l.client.Redis.HGet(ctx, l.client.GenKey(common.ChannelKey, guildId, channelId), "id").Result()
			if err != nil {
				log.Fatalf("[%v] Couldn't perform HGET: %v", l.ListenerInfo().Event, err)
			}
			if channel != "" {
				parsed, _ := json.Marshal(data.LastPinTimestamp)
				if _, err := l.client.Redis.HSet(ctx, l.client.GenKey(common.ChannelKey, guildId, channelId), "last_pin_timestamp", string(parsed)).Result(); err != nil {
					log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
				}
			}
		} else {
			channel := l.client.Redis.HGet(context.Background(), l.client.GenKey(common.ChannelKey, channelId), "id").String()
			if channel != "" {
				parsed, _ := json.Marshal(data.LastPinTimestamp)
				if _, err := l.client.Redis.HSet(ctx, l.client.GenKey(common.ChannelKey, channelId), "last_pin_timestamp", string(parsed)).Result(); err != nil {
					log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
				}

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

func (l ChannelPinsUpdateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeChannelPinsUpdate,
	}
}

func RegisterChannelPinsUpdateListener(client *lib.GatewayClient) {
	common.RegisterListener(&ChannelPinsUpdateListener{*client})
}
