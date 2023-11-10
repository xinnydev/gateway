package listener

import (
	"context"
	"encoding/json"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type ChannelDeleteListener struct {
	client lib.GatewayClient
}

func (l ChannelDeleteListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventChannelDelete)
	ctx := context.Background()
	channelId := data.ID().String()
	if *l.client.Config.State.Channel {
		if data.Type() == discord.ChannelTypeDM {
			if _, err := l.client.Redis.SRem(ctx, l.client.GenKey(common.ChannelKey, common.KeysSuffix), channelId).Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SREM: %v", l.ListenerInfo().Event, err)
			}
			if err := l.client.Redis.Unlink(ctx, l.client.GenKey(common.ChannelKey, channelId)); err != nil {
				log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
			}
		} else {
			guildId := data.GuildID().String()
			if _, err := l.client.Redis.SRem(
				context.Background(),
				l.client.GenKey(common.ChannelKey, common.KeysSuffix, guildId),
				channelId).Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform SREM: %v", l.ListenerInfo().Event, err)
			}
			if err := l.client.Redis.Unlink(ctx, l.client.GenKey(common.ChannelKey, common.KeysSuffix, guildId, channelId)); err != nil {
				log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
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

func (l ChannelDeleteListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeChannelDelete,
	}
}

func RegisterChannelDeleteListener(client *lib.GatewayClient) {
	common.RegisterListener(&ChannelDeleteListener{*client})
}
