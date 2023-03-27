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
	listener.RegisterChannelCreateListener(gateway)
	listener.RegisterChannelDeleteListener(gateway)
	listener.RegisterChannelPinsUpdateListener(gateway)
	listener.RegisterGuildCreateListener(gateway)
	listener.RegisterReadyListener(gateway)
	listener.RegisterMessageCreateListener(gateway)
	select {}
}
