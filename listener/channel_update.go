package listener

import (
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type ChannelUpdateListener struct {
	client lib.GatewayClient
}

func (l ChannelUpdateListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventChannelUpdate)
	if *l.client.Config.State.Channel {
		if data.Type() == discord.ChannelTypeDM {
			if _, err := l.client.Redis.Hset(fmt.Sprintf("%v:%v", common.ChannelKey, data.ID()), data); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		} else {
			if _, err := l.client.Redis.Hset(fmt.Sprintf("%v:%v:%v", common.ChannelKey, data.GuildID(), data.ID()), data); err != nil {
				log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	body, _ := data.MarshalJSON()
	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
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
