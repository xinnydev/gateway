package lib

import (
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/config"
	"github.com/xinny/gateway/redis"
	"time"
)

type GatewayClient struct {
	Broker         *broker.Broker
	Rest           rest.Rest
	Ws             gateway.Gateway
	BotApplication *discord.Application
	Redis          *redis.Client
	Config         *config.Config
	startTime      time.Time
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

	b := broker.NewBroker(user.ID.String(), *conf.AMQPUrl)
	_, err = b.Channel.QueueDeclare(common.Exchange, false, false, false, false, nil)
	err = b.Channel.ExchangeDeclare(user.ID.String(), "direct", false, false, false, false, nil)
	if err != nil {
		panic(fmt.Sprintf("couldn't declare amqp topic: %v", err))
	}

	client.Broker = b
	client.Ws = gateway.New(*conf.DiscordToken, client.handleWsEvent, handleWsClose, func(c *gateway.Config) {
		if conf.Gateway.HandshakeTimeout != nil {
			c.Dialer.HandshakeTimeout = time.Duration(*conf.Gateway.HandshakeTimeout)
		}
		if conf.Gateway.ShardCount != nil {
			c.ShardCount = *conf.Gateway.ShardCount
		}
	},
		gateway.WithIntents(*conf.Gateway.Intents),
		gateway.WithCompress(true),
		gateway.WithLargeThreshold(*conf.Gateway.LargeThreshold))

	client.BotApplication = user
	client.startTime = time.Now()
	return &client
}

func handleWsClose(gateway gateway.Gateway, err error) {

}

func (c *GatewayClient) handleWsEvent(gatewayEventType gateway.EventType, sequenceNumber int, shardID int, event gateway.EventData) {
	for _, listener := range common.Listeners {
		if listener.ListenerInfo().Event == gatewayEventType {
			listener.Run(event)
			break
		}
	}
}
