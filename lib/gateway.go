package lib

import (
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
	"time"
)

type GatewayClient struct {
	Broker         *broker.Broker
	Rest           rest.Rest
	ShardManager   sharding.ShardManager
	BotApplication *discord.Application
	Redis          *redis.Client
	Config         *config.Config
}

func NewGateway(conf config.Config) *GatewayClient {

	client := GatewayClient{
		Rest:   rest.New(rest.NewClient(*conf.DiscordToken)),
		Redis:  redis.NewRedisClient(conf.Redis),
		Config: &conf,
	}

	// Fetch client user
	user, err := client.Rest.GetBotApplicationInfo()

	if err != nil {
		panic(fmt.Sprintf("couldn't fetch self-user: %v", err))
	}
	client.BotApplication = user

	b := broker.NewBroker(user.ID.String(), *conf.AMQPUrl)
	_, err = b.Channel.QueueDeclare(common.Exchange, false, false, false, false, nil)
	err = b.Channel.ExchangeDeclare(user.ID.String(), "direct", false, false, false, false, nil)
	if err != nil {
		panic(fmt.Sprintf("couldn't declare amqp topic: %v", err))
	}

	client.Broker = b
	client.ShardManager = sharding.New(*conf.DiscordToken, client.handleWsEvent, sharding.WithGatewayConfigOpts(
		gateway.WithIntents(*conf.Gateway.Intents),
		gateway.WithCompress(true),
		gateway.WithEnableRawEvents(true),
		gateway.WithLargeThreshold(*conf.Gateway.LargeThreshold), func(c *gateway.Config) {
			if conf.Gateway.HandshakeTimeout != nil {
				c.Dialer.HandshakeTimeout = time.Duration(*conf.Gateway.HandshakeTimeout)
			}
		}), sharding.WithShardCount(*conf.Gateway.ShardCount), sharding.WithGatewayCreateFunc(func(token string, eventHandlerFunc gateway.EventHandlerFunc, closeHandlerFUnc gateway.CloseHandlerFunc, opts ...gateway.ConfigOpt) gateway.Gateway {
		g := gateway.New(token, eventHandlerFunc, closeHandlerFUnc, opts...)
		var sessionData redis.SessionData
		exists, err := client.Redis.HGetAllAndParse(
			fmt.Sprintf("%v:%v:%v", common.SessionKey, client.BotApplication.ID.String(), g.ShardID()),
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
