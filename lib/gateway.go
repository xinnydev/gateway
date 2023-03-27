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
)

type GatewayClient struct {
	Broker         *broker.Broker
	Rest           rest.Rest
	Ws             gateway.Gateway
	BotApplication *discord.Application
	Redis          *redis.Client
	Config         *config.Config
}

func NewGateway(gatewayConf config.Config) *GatewayClient {
	client := GatewayClient{
		Rest:   rest.New(rest.NewClient(*gatewayConf.DiscordToken)),
		Redis:  redis.NewRedisClient(gatewayConf.Redis),
		Config: &gatewayConf,
	}

	// Fetch client user
	user, err := client.Rest.GetBotApplicationInfo()

	if err != nil {
		panic(fmt.Sprintf("couldn't fetch self-user: %v", err))
	}

	b := broker.NewBroker(user.ID.String(), *gatewayConf.AMQPUrl)

	client.Broker = b
	client.Ws = gateway.New(*gatewayConf.DiscordToken, client.handleWsEvent, handleWsClose, gateway.WithIntents(131071))
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
