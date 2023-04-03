package listener

import (
	"bytes"
	"context"
	"fmt"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/json"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type RawListener struct {
	client lib.GatewayClient
}

func (l RawListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventRaw)
	if l.client.ShardManager.Shard(shardID).LastSequenceReceived() != nil {
		if _, err := l.client.Redis.HSet(context.Background(),
			fmt.Sprintf("%v:%v:%v", common.SessionKey, l.client.BotID, shardID), "last_seq",
			*l.client.ShardManager.Shard(shardID).LastSequenceReceived()).Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
		}
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(data.Payload)
	eventData, _ := gateway.UnmarshalEventData(buf.Bytes(), data.EventType)

	switch data.EventType {
	case gateway.EventTypeChannelCreate:
	case gateway.EventTypeChannelDelete:
	case gateway.EventTypeChannelPinsUpdate:
	case gateway.EventTypeChannelUpdate:
	case gateway.EventTypeGuildCreate:
	case gateway.EventTypeGuildDelete:
	case gateway.EventTypeGuildEmojisUpdate:
	case gateway.EventTypeGuildMemberAdd:
	case gateway.EventTypeGuildMemberRemove:
	case gateway.EventTypeGuildMemberUpdate:
	case gateway.EventTypeGuildMembersChunk:
	case gateway.EventTypeGuildRoleCreate:
	case gateway.EventTypeGuildRoleDelete:
	case gateway.EventTypeGuildRoleUpdate:
	case gateway.EventTypeGuildUpdate:
	case gateway.EventTypeHeartbeatAck:
	case gateway.EventTypeMessageCreate:
	case gateway.EventTypeMessageDelete:
	case gateway.EventTypeMessageDeleteBulk:
	case gateway.EventTypeReady:
	case gateway.EventTypeUserUpdate:
	case gateway.EventTypeVoiceStateUpdate:
		rawEventHandler, ok := common.Listeners[string(data.EventType)]
		if ok {
			rawEventHandler.Run(shardID, eventData)
		}
		break
	default:
		body, _ := json.Marshal(eventData)
		if err := l.client.Broker.Publish(string(data.EventType), body); err != nil {
			log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
			return
		}
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
