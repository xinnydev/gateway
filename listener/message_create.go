package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/streadway/amqp"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
	"github.com/xinny/gateway/redis"
)

type MessageCreateListener struct {
	client lib.GatewayClient
}

func (l MessageCreateListener) Run(ev gateway.EventData) {
	ctx := context.Background()
	data := ev.(gateway.EventMessageCreate)
	mappedData, err := redis.StructToMap(data)
	if err != nil {
		log.Fatalf("[%v] Couldn't convert struct to map: %v", l.ListenerInfo().Event, err)
	}
	stringifiedData := redis.IterateMapAndStringify(mappedData)
	if *l.client.Config.State.Message {
		_, err = l.client.Redis.SAdd(ctx, fmt.Sprintf("%v%v", common.MessageKey, common.KeysSuffix), data.ID.String()).Result()
		if err != nil {
			log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
		}
		_, err = l.client.Redis.HSet(ctx, fmt.Sprintf("%v:%v", common.MessageKey, data.ID.String()), stringifiedData).Result()
		if err != nil {
			log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
		}

	}
	body, _ := json.Marshal(data)
	err = l.client.BrokerChannel.Publish(l.client.BotApplication.ID.String(), string(l.ListenerInfo().Event), false, false, amqp.Publishing{
		Body: body,
	})
	if err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}
}

func (l MessageCreateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeMessageCreate,
	}
}

func RegisterMessageCreateListener(client *lib.GatewayClient) {
	common.RegisterListener(&MessageCreateListener{*client})
}
