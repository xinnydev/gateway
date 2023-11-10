package listener

import (
	"bytes"
	"context"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/json"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
	"strconv"
)

type RawListener struct {
	client lib.GatewayClient
}

func (l RawListener) Run(shardID int, ev gateway.EventData) {
	data := ev.(gateway.EventRaw)
	if l.client.ShardManager.Shard(shardID).LastSequenceReceived() != nil {
		if _, err := l.client.Redis.HSet(context.Background(),
			l.client.GenKey(common.SessionKey, strconv.Itoa(shardID)), "last_seq",
			*l.client.ShardManager.Shard(shardID).LastSequenceReceived()).Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
		}
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(data.Payload)
	if err != nil {
		log.Errorf("[%v] Couldn't parse payload: %v", l.ListenerInfo().Event, err)
		return
	}
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
		body, _ := json.Marshal(&broker.PublishPayload{
			ShardID: shardID,
			Data:    data,
		})
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
