package lib

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/disgoorg/json"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/config"
	"github.com/xinny/gateway/redis"
	"github.com/xinny/gateway/utils"
	"strconv"
	"strings"
	"time"
)

func NewGateway(conf config.Config) IGatewayClient {
	clientId, _ := base64.StdEncoding.DecodeString(strings.Split(*conf.DiscordToken, ".")[0])
	client := GatewayClient{
		BotID:  string(clientId),
		Broker: broker.NewBroker(string(clientId), *conf.AMQPUrl),
		Redis:  redis.NewRedisClient(conf.Redis),
		Config: &conf,
	}

	// Check for session
	sessionKeys, _ := client.Redis.ScanKeys(fmt.Sprintf("%v:%v:*", common.SessionKey, client.BotID))

	if len(sessionKeys) == 0 {
		//client.Redis.ClearCache()
	}

	// Declare the queue and the exchange
	_, err := client.Broker.Channel.QueueDeclare(common.Exchange, false, false, false, false, nil)
	err = client.Broker.Channel.ExchangeDeclare(client.BotID, "direct", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("couldn't declare amqp topic: %v", err)
	}

	err = client.Broker.Channel.QueueBind(common.Exchange, "send", client.BotID, false, nil)
	if err != nil {
		log.Fatalf("couldn't bind amqp queue: %v", err)
	}

	consumer, err := client.Broker.Channel.Consume(common.Exchange, "send", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("error consuming queue: %v", err)
	}

	go func() {
		for v := range consumer {
			var payload gateway.Message
			err = json.Unmarshal(v.Body, &payload)
			log.Infof("[%v] received ws send request with opcode: %v", common.Exchange, payload.Op)
			if err != nil {
				log.Warnf("[%v] couldn't unmarshal received ws consumer: %v", common.Exchange, err)
				return
			}
			err := client.ShardManager.Shard(payload.S).Send(context.Background(), payload.Op, payload.D)
			if err != nil {
				log.Warnf("[%v] couldn't forward received payload: %v", common.Exchange, err)
				return
			}
		}
	}()

	// Calculate shard ids
	var shardIds []int
	if conf.Gateway.ShardStart != nil && conf.Gateway.ShardEnd != nil {
		for i := *conf.Gateway.ShardStart; i <= *conf.Gateway.ShardEnd; i++ {
			shardIds = append(shardIds, i)
		}
	}

	// Declare ShardManager
	client.ShardManager = sharding.New(*conf.DiscordToken, client.handleWsEvent,
		sharding.WithShardCount(*conf.Gateway.ShardCount),
		sharding.WithAutoScaling(true),
		sharding.WithGatewayCreateFunc(client.createGatewayFunc),
		sharding.WithShardIDs(shardIds...),
		sharding.WithGatewayConfigOpts(
			gateway.WithEnableRawEvents(true),
			gateway.WithIntents(*conf.Gateway.Intents),
			gateway.WithCompress(true),
			gateway.WithLargeThreshold(*conf.Gateway.LargeThreshold),
			gateway.WithPresenceOpts(gateway.WithOnlineStatus(*conf.Gateway.Presence.Status)),
			func(gConf *gateway.Config) {
				if conf.Gateway.HandshakeTimeout != nil {
					gConf.Dialer.HandshakeTimeout = time.Duration(*conf.Gateway.HandshakeTimeout) * time.Millisecond
				}
				if conf.Gateway.Presence.Type != nil && conf.Gateway.Presence.Name != nil {
					gConf.Presence.Activities = []discord.Activity{
						{
							Name: *conf.Gateway.Presence.Name,
							Type: discord.ActivityType(*conf.Gateway.Presence.Type),
						},
					}
				}
			},
		))

	return &client
}

func (c *GatewayClient) handleWsEvent(gatewayEventType gateway.EventType, sequenceNumber int, shardID int, event gateway.EventData) {
	if gatewayEventType == gateway.EventTypeRaw {
		rawEventHandler, ok := common.Listeners[string(gatewayEventType)]
		if ok {
			rawEventHandler.Run(shardID, event)
		}
	}
}

func (c *GatewayClient) createGatewayFunc(token string, eventHandler gateway.EventHandlerFunc, closeHandler gateway.CloseHandlerFunc, opts ...gateway.ConfigOpt) gateway.Gateway {
	g := gateway.New(token, eventHandler, closeHandler, opts...)
	var sessionData redis.SessionData
	exists, err := c.Redis.HGetAllAndParse(
		c.GenKey(common.SessionKey, strconv.Itoa(g.ShardID())),
		&sessionData)

	if err != nil {
		log.Fatalf("unable to fetch session data: %v", err)
	}

	if !exists {
		return g
	}

	return gateway.New(token, eventHandler, closeHandler, func(config *gateway.Config) {
		for _, v := range opts {
			v(config)
		}
		config.SessionID = &sessionData.SessionID
		config.ResumeURL = &sessionData.ResumeURL
		config.EnableResumeURL = true
		log.Debugf("[%v/%v] resuming session: %v", config.ShardID, config.ShardCount, sessionData.SessionID)
	})
}

func (c *GatewayClient) Client() *GatewayClient {
	return c
}

func (c *GatewayClient) GenKey(str ...string) string {
	if len(str) > 1 && str[1] == common.KeysSuffix {
		str[0] = fmt.Sprintf("%s%s", str[0], common.KeysSuffix)
		str[1] = ""
	}
	str = append([]string{c.BotID}, str...)

	var sanitized []string
	for _, s := range str {
		if s != "" {
			sanitized = append(sanitized, s)
		}
	}

	return utils.GenerateKey(sanitized...)
}
