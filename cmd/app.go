package cmd

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/xinny/gateway/config"
	"github.com/xinny/gateway/lib"
	"github.com/xinny/gateway/listener"
)

func Run() {
	var err error
	if err = godotenv.Load(); err != nil {
		panic(fmt.Sprintf("couldn't load .env file: %v", err))
	}

	conf, err := config.Init()
	if err != nil {
		panic(fmt.Sprintf("couldn't initialize config: %v", err))
	}

	ctx := context.Background()
	gateway := lib.NewGateway(conf)
	if err = gateway.Ws.Open(ctx); err != nil {
		panic(fmt.Sprintf("couldn't open ws connection: %v", err))
	}
	// TODO: Impl ready timeout handler
	// TODO: Impl hello timeout handler
	// TODO: Impl sharding

	// Register all listeners
	listener.RegisterChannelCreateListener(gateway)
	listener.RegisterChannelDeleteListener(gateway)
	listener.RegisterChannelPinsUpdateListener(gateway)
	listener.RegisterChannelUpdateListener(gateway)
	listener.RegisterGuildCreateListener(gateway)
	listener.RegisterGuildDeleteListener(gateway)
	listener.RegisterGuildEmojisUpdateListener(gateway)
	listener.RegisterGuildMemberAddListener(gateway)
	listener.RegisterGuildMemberRemoveListener(gateway)
	listener.RegisterGuildMemberUpdateListener(gateway)
	listener.RegisterGuildMembersChunkListener(gateway)
	listener.RegisterGuildRoleCreateListener(gateway)
	listener.RegisterGuildRoleDeleteListener(gateway)
	listener.RegisterGuildRoleUpdateListener(gateway)
	listener.RegisterMessageCreateListener(gateway)
	listener.RegisterMessageDeleteListener(gateway)
	listener.RegisterMessageDeleteBulkListener(gateway)
	listener.RegisterReadyListener(gateway)
	listener.RegisterUserUpdateListener(gateway)
	listener.RegisterVoiceStateUpdateListener(gateway)
	select {}
}
