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
)

type ChannelDeleteListener struct {
	client lib.GatewayClient
}

func (l ChannelDeleteListener) Run(ev gateway.EventData) {
	data := ev.(gateway.EventChannelDelete)
	ctx := context.Background()
	if *l.client.Config.State.Channel {
		if data.Type() == discord.ChannelTypeDM {
			if _, err := l.client.Redis.SRem(ctx, fmt.Sprintf("%v%v", common.ChannelKey, common.KeysSuffix), data.ID()).Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SREM: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.Unlink(ctx, fmt.Sprintf("%v:%v", common.ChannelKey, data.ID())).Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
			}
		} else {
			if _, err := l.client.Redis.SRem(
				ctx,
				fmt.Sprintf("%v%v", common.ChannelKey, common.KeysSuffix),
				fmt.Sprintf("%v:%v", data.GuildID(), data.ID())).Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SREM: %v", l.ListenerInfo().Event, err)
			}
			if _, err := l.client.Redis.Unlink(ctx, fmt.Sprintf("%v:%v:%v", common.ChannelKey, data.GuildID(), data.ID())).Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
			}
		}
	}
	body, _ := data.MarshalJSON()
	if err := l.client.BrokerChannel.Publish(l.client.BotApplication.ID.String(), string(l.ListenerInfo().Event), false, false, amqp.Publishing{
		Body: body,
	}); err != nil {
		log.Errorf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
	}
	return
}

func (l ChannelDeleteListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeChannelDelete,
	}
}

func RegisterChannelDeleteListener(client *lib.GatewayClient) {
	common.RegisterListener(&ChannelDeleteListener{*client})
}
