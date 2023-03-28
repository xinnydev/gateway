package config

import (
	"github.com/caarlos0/env/v7"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/joho/godotenv"
	"github.com/xinny/gateway/redis"
)

type Config struct {
	DiscordToken *string `env:"DISCORD_TOKEN,required"`
	Gateway      struct {
		Intents          *gateway.Intents `env:"GATEWAY_INTENTS,required"`
		HelloTimeout     *int             `env:"GATEWAY_HELLO_TIMEOUT"`
		ReadyTimeout     *int             `env:"GATEWAY_READY_TIMEOUT"`
		HandshakeTimeout *int             `env:"GATEWAY_HANDSHAKE_TIMEOUT"`
		LargeThreshold   *int             `env:"GATEWAY_LARGE_THRESHOLD" envDefault:"250"`
		ShardCount       *int             `env:"GATEWAY_SHARD_COUNT"`
		ShardStart       *int             `env:"GATEWAY_SHARD_START"`
		ShardEnd         *int             `env:"GATEWAY_SHARD_END"`
	}
	AMQPUrl *string `env:"AMQP_URL,required"`
	Redis   redis.Config
	State   struct {
		Channel      *bool `env:"STATE_CHANNEL,required"`
		Emoji        *bool `env:"STATE_EMOJI,required"`
		Member       *bool `env:"STATE_MEMBER,required"`
		Message      *bool `env:"STATE_MESSAGE,required"`
		Presence     *bool `env:"STATE_PRESENCE,required"`
		Reaction     *bool `env:"STATE_REACTION,required"`
		Role         *bool `env:"STATE_ROLE,required"`
		Sticker      *bool `env:"STATE_STICKER,required"`
		Thread       *bool `env:"STATE_THREAD,required"`
		ThreadMember *bool `env:"STATE_THREAD_MEMBER,required"`
		User         *bool `env:"STATE_USER,required"`
		Voice        *bool `env:"STATE_VOICE,required"`
	}
}

func Init() (conf Config, err error) {
	err = godotenv.Load()
	if err != nil {
		return
	}
	conf = Config{}
	if err := env.Parse(&conf); err != nil {
		log.Fatalf("%+v\n", err)
	}
	return conf, nil
}
