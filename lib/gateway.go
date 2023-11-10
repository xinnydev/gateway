package lib

import (
	"github.com/disgoorg/disgo/sharding"
	"github.com/xinny/gateway/broker"
	"github.com/xinny/gateway/config"
	"github.com/xinny/gateway/redis"
)

type GatewayClient struct {
	BotID        string
	Broker       *broker.Broker
	ShardManager sharding.ShardManager
	Redis        *redis.Client
	Config       *config.Config
}

type IGatewayClient interface {
	Client() *GatewayClient
	GenKey(string ...string) string
}
