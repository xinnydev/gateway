package lib

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/sharding"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/config"
	"github.com/xinny/gateway/redis"
	"strings"
)

type GatewayClient struct {
	BotID        string
	Broker       *broker.Broker
	Rest         rest.Rest
	ShardManager sharding.ShardManager
	Redis        *redis.Client
	Config       *config.Config
}

func NewGateway(conf config.Config) *GatewayClient {
	clientId, _ := base64.StdEncoding.DecodeString(strings.Split(*conf.DiscordToken, ".")[0])
	client := GatewayClient{
		BotID:  string(clientId),
		Broker: broker.NewBroker(string(clientId), *conf.AMQPUrl),
		Rest:   rest.New(rest.NewClient(*conf.DiscordToken)),
		Redis:  redis.NewRedisClient(conf.Redis),
		Config: &conf,
	}

	// Check for session
	var sessionKeys []string
	sessionKeysIterator := client.Redis.Scan(context.Background(), 0, fmt.Sprintf("%v:%v:*", common.SessionKey, client.BotID), 0).Iterator()
	for sessionKeysIterator.Next(context.Background()) {
		sessionKeys = append(sessionKeys, sessionKeysIterator.Val())
	}

	if len(sessionKeys) == 0 {
		client.Redis.ClearCache()
	}

	// Declare the queue and the exchange
	_, err := client.Broker.Channel.QueueDeclare(common.Exchange, false, false, false, false, nil)
	err = client.Broker.Channel.ExchangeDeclare(client.BotID, "direct", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("couldn't declare amqp topic: %v", err)
	}

	// Declare ShardManager
	client.ShardManager = sharding.New(*conf.DiscordToken, client.handleWsEvent,
		sharding.WithGatewayConfigOpts(
			func(config *gateway.Config) {
				config.Intents = *conf.Gateway.Intents
				config.EnableRawEvents = true
				config.Compress = true
				config.LargeThreshold = *conf.Gateway.LargeThreshold
				config.Presence = &gateway.MessageDataPresenceUpdate{
					Status: *conf.Gateway.Presence.Status,
				}

				if conf.Gateway.Presence.Type != nil && conf.Gateway.Presence.Name != nil {
					config.Presence.Activities = []discord.Activity{
						{
							Name: *conf.Gateway.Presence.Name,
							Type: discord.ActivityType(*conf.Gateway.Presence.Type),
						},
					}
				}
			}), func(shardConf *sharding.Config) {
			shardConf.GatewayCreateFunc = client.createGatewayFunc
			shardConf.ShardCount = *conf.Gateway.ShardCount
			if conf.Gateway.ShardStart != nil && conf.Gateway.ShardEnd != nil {
				shardConf.ShardIDs = map[int]struct{}{}
				for i := *conf.Gateway.ShardStart; i <= *conf.Gateway.ShardEnd; i++ {
					shardConf.ShardIDs[i] = struct{}{}
				}
			}
		})

	return &client
}

func (c *GatewayClient) handleWsEvent(gatewayEventType gateway.EventType, sequenceNumber int, shardID int, event gateway.EventData) {
	for _, listener := range common.Listeners {
		if listener.ListenerInfo().Event == gatewayEventType {
			listener.Run(event)
			break
		}
	}
}

func (c *GatewayClient) createGatewayFunc(token string, eventHandler gateway.EventHandlerFunc, closeHandler gateway.CloseHandlerFunc, opts ...gateway.ConfigOpt) gateway.Gateway {
	options := gateway.Config{}
	for _, opt := range opts {
		opt(&options)
	}

	var sessionData redis.SessionData
	exists, err := c.Redis.HGetAllAndParse(
		fmt.Sprintf("%v:%v:%v", common.SessionKey, c.BotID, options.ShardID),
		&sessionData)
	if err != nil {
		log.Fatalf("unable to fetch session data: %v", err)
	}

	if exists {
		options.SessionID = &sessionData.SessionID
		options.ResumeURL = &sessionData.ResumeURL
		options.EnableResumeURL = true
		log.Debugf("[%v/%v] resuming session: %v", options.ShardID, options.ShardCount, sessionData.SessionID)
	}

	return gateway.New(token, eventHandler, closeHandler, func(config *gateway.Config) {
		config = &options
	})
}
