package listener

import (
	"context"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/streadway/amqp"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
	"github.com/xinny/gateway/redis"
)

type ChannelUpdateListener struct {
	client lib.GatewayClient
}

func (l ChannelUpdateListener) Run(ev gateway.EventData) {
	data := ev.(gateway.EventChannelUpdate)
	ctx := context.Background()
	if *l.client.Config.State.Channel {
		mappedData, err := redis.StructToMap(data)
		if err != nil {
			log.Fatalf("[%v] Couldn't convert struct to map: %v", l.ListenerInfo().Event, err)
		}
		stringifiedData := redis.IterateMapAndStringify(mappedData)

		if data.Type() == discord.ChannelTypeDM {
			if _, err = l.client.Redis.HSet(ctx, fmt.Sprintf("%v:%v", common.ChannelKey, data.ID()), stringifiedData).Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		} else {
			if _, err = l.client.Redis.HSet(ctx, fmt.Sprintf("%v:%v:%v", common.ChannelKey, data.GuildID(), data.ID()), stringifiedData).Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	body, _ := data.MarshalJSON()
	err := l.client.BrokerChannel.Publish(l.client.BotApplication.ID.String(), string(l.ListenerInfo().Event), false, false, amqp.Publishing{
		Body: body,
	})
	if err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}
}

func (l ChannelUpdateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeChannelUpdate,
	}
}

func RegisterChannelUpdateListener(client *lib.GatewayClient) {
	common.RegisterListener(&ChannelUpdateListener{*client})
}
