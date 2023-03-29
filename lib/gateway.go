package lib

import (
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
	"time"
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
		Rest:   rest.New(rest.NewClient(*conf.DiscordToken)),
		Redis:  redis.NewRedisClient(conf.Redis),
		Config: &conf,
	}

	b := broker.NewBroker(client.BotID, *conf.AMQPUrl)
	_, err := b.Channel.QueueDeclare(common.Exchange, false, false, false, false, nil)
	err = b.Channel.ExchangeDeclare(client.BotID, "direct", false, false, false, false, nil)
	if err != nil {
		log.Fatalf(fmt.Sprintf("couldn't declare amqp topic: %v", err))
	}

	sessionKeys, err := client.Redis.ScanKeys(fmt.Sprintf("%v*", common.SessionKey))
	if err != nil {
		log.Fatalf("couldn't fetch session keys: %v", err)
	}
	if len(sessionKeys) < 1 {
		client.Redis.ClearCache()
	}
	client.Broker = b
	client.ShardManager = sharding.New(*conf.DiscordToken, client.handleWsEvent, sharding.WithGatewayConfigOpts(
		gateway.WithIntents(*conf.Gateway.Intents),
		gateway.WithCompress(true),
		gateway.WithPresenceOpts(func(p *gateway.MessageDataPresenceUpdate) {
			p.Status = *conf.Gateway.Presence.Status
			if conf.Gateway.Presence.Type != nil && conf.Gateway.Presence.Name != nil {
				p.Activities = []discord.Activity{
					{
						Name: *conf.Gateway.Presence.Name,
						Type: discord.ActivityType(*conf.Gateway.Presence.Type),
					},
				}
			}
		}),
		gateway.WithLargeThreshold(*conf.Gateway.LargeThreshold), func(c *gateway.Config) {
			if conf.Gateway.HandshakeTimeout != nil {
				c.Dialer.HandshakeTimeout = time.Duration(*conf.Gateway.HandshakeTimeout)
			}
		}), sharding.WithShardCount(*conf.Gateway.ShardCount), sharding.WithGatewayCreateFunc(func(token string, eventHandlerFunc gateway.EventHandlerFunc, closeHandlerFUnc gateway.CloseHandlerFunc, opts ...gateway.ConfigOpt) gateway.Gateway {
		g := gateway.New(token, eventHandlerFunc, closeHandlerFUnc, opts...)
		var sessionData redis.SessionData
		exists, err := client.Redis.HGetAllAndParse(
			fmt.Sprintf("%v:%v:%v", common.SessionKey, client.BotID, g.ShardID()),
			&sessionData)
		if err != nil {
			log.Fatalf("unable to fetch session data: %v", err)
		}
		if !exists {
			return g
		}
		g = gateway.New(token, eventHandlerFunc, closeHandlerFUnc, func(config *gateway.Config) {
			for _, v := range opts {
				v(config)
			}
			config.SessionID = &sessionData.SessionID
			config.ResumeURL = &sessionData.ResumeURL
			config.EnableResumeURL = true
			log.Debugf("[%v/%v] resuming session: %v", config.ShardID, config.ShardCount, sessionData.SessionID)
		})
		return g
	}), func(config *sharding.Config) {
		if conf.Gateway.ShardStart != nil && conf.Gateway.ShardEnd != nil {
			config.ShardIDs = map[int]struct{}{}
			for i := *conf.Gateway.ShardStart; i <= *conf.Gateway.ShardEnd; i++ {
				config.ShardIDs[i] = struct{}{}
			}
		}
	})

	return &client
}

func (c *GatewayClient) handleWsEvent(gatewayEventType gateway.EventType, sequenceNumber int, shardID int, event gateway.EventData) {
	//log.Infof("%v", gatewayEventType)
	for _, listener := range common.Listeners {
		if listener.ListenerInfo().Event == gatewayEventType {
			listener.Run(event)
			break
		}
	}
}
