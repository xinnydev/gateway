package cmd

import (
	"context"
	"fmt"
	"github.com/disgoorg/log"
	"github.com/joho/godotenv"
	"github.com/xinny/gateway/config"
	"github.com/xinny/gateway/lib"
	"github.com/xinny/gateway/listener"
)

func Run() {
	var err error
	_ = godotenv.Load()

	conf, err := config.Init()
	if err != nil {
		panic(fmt.Sprintf("couldn't initialize config: %v", err))
	}

	ctx := context.Background()
	gateway := lib.NewGateway(conf)
	log.SetLevel(log.LevelDebug)

	gatewayClient := gateway.Client()
	// Register all listeners
	listener.RegisterRawListener(gatewayClient)
	listener.RegisterHeartbeatAckListener(gatewayClient)
	listener.RegisterChannelCreateListener(gatewayClient)
	listener.RegisterChannelDeleteListener(gatewayClient)
	listener.RegisterChannelPinsUpdateListener(gatewayClient)
	listener.RegisterChannelUpdateListener(gatewayClient)
	listener.RegisterGuildCreateListener(gatewayClient)
	listener.RegisterGuildDeleteListener(gatewayClient)
	listener.RegisterGuildEmojisUpdateListener(gatewayClient)
	listener.RegisterGuildMemberAddListener(gatewayClient)
	listener.RegisterGuildMemberRemoveListener(gatewayClient)
	listener.RegisterGuildMemberUpdateListener(gatewayClient)
	listener.RegisterGuildMembersChunkListener(gatewayClient)
	listener.RegisterGuildRoleCreateListener(gatewayClient)
	listener.RegisterGuildRoleDeleteListener(gatewayClient)
	listener.RegisterGuildRoleUpdateListener(gatewayClient)
	listener.RegisterGuildUpdateListener(gatewayClient)
	listener.RegisterMessageCreateListener(gatewayClient)
	listener.RegisterMessageDeleteListener(gatewayClient)
	listener.RegisterMessageDeleteBulkListener(gatewayClient)
	listener.RegisterReadyListener(gatewayClient)
	listener.RegisterUserUpdateListener(gatewayClient)
	listener.RegisterVoiceStateUpdateListener(gatewayClient)

	gatewayClient.ShardManager.Open(ctx)
	select {}
}
