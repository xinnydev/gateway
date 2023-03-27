package lib

import (
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/streadway/amqp"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/config"
	"github.com/xinny/gateway/redis"
)

type GatewayClient struct {
	BrokerChannel  *amqp.Channel
	Rest           rest.Rest
	Ws             gateway.Gateway
	BotApplication *discord.Application
	Redis          *redis.RedisClient
	Config         *config.Config
}

func NewGateway(gatewayConf config.Config) *GatewayClient {
	b := broker.NewBroker(*gatewayConf.AMQPUrl)
	client := GatewayClient{
		BrokerChannel: b.Channel,
		Rest:          rest.New(rest.NewClient(*gatewayConf.DiscordToken)),
		Redis:         redis.NewRedisClient(gatewayConf.Redis),
		Config:        &gatewayConf,
	}

	client.Ws = gateway.New(*gatewayConf.DiscordToken, client.handleWsEvent, handleWsClose, gateway.WithIntents(131071))

	// Fetch client user
	user, err := client.Rest.GetBotApplicationInfo()
	if err != nil {
		panic(fmt.Sprintf("couldn't fetch self-user: %v", err))
	}

	client.BotApplication = user
	err = b.Channel.ExchangeDeclare(user.ID.String(), "topic", false, false, false, false, nil)
	if err != nil {
		panic(fmt.Sprintf("couldn't declare amqp topic: %v", err))
	}

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
