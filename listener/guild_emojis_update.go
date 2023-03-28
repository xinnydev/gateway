package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
	"strings"
)

type GuildEmojisUpdateListener struct {
	client lib.GatewayClient
}

func (l GuildEmojisUpdateListener) Run(ev gateway.EventData) {
	ctx := context.Background()
	data := ev.(gateway.EventGuildEmojisUpdate)
	var emojis []discord.Emoji
	keys, err := l.client.Redis.SMembers(ctx, fmt.Sprintf("%v%v", common.EmojiKey, common.KeysSuffix)).Result()
	if err != nil {
		log.Fatalf("[%v] Couldn't perform SMEMBERS: %v", l.ListenerInfo().Event, err)
	}

	for _, v := range keys {
		var emoji discord.Emoji
		exists, err := l.client.Redis.HGetAllAndParse(fmt.Sprintf("%v:%v", common.EmojiKey, v), &emoji)
		if err != nil {
			log.Fatalf("[%v] Couldn't perform HGetAllAndParse: %v", l.ListenerInfo().Event, err)
		}

		if !exists {
			continue
		}

		for _, emoji := range data.Emojis {
			if emoji.ID.String() == emoji.ID.String() {
				if _, err := l.client.Redis.
					SRem(ctx, fmt.Sprintf("%v%v", common.EmojiKey, common.KeysSuffix), v).
					Result(); err != nil {
					log.Fatalf("[%v] Couldn't perform SREM: %v", l.ListenerInfo().Event, err)
				}

				if _, err := l.client.Redis.
					Unlink(ctx, fmt.Sprintf("%v:%v", common.EmojiKey, v)).
					Result(); err != nil {
					log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
				}

				emojis = append(emojis, emoji)
			}
		}

		if strings.HasPrefix(v, data.GuildID.String()) {
			if _, err := l.client.Redis.
				Unlink(ctx, fmt.Sprintf("%v:%v", common.EmojiKey, v)).
				Result(); err != nil {
				log.Fatalf("[%v] Couldn't perform UNLINK: %v", l.ListenerInfo().Event, err)
			}
		}
	}

	for _, v := range data.Emojis {
		if _, err := l.client.Redis.
			SAdd(ctx,
				fmt.Sprintf("%v%v", common.EmojiKey, common.KeysSuffix),
				fmt.Sprintf("%v:%v", data.GuildID.String(), v.ID.String())).
			Result(); err != nil {
			log.Fatalf("[%v] Couldn't perform SADD: %v", l.ListenerInfo().Event, err)
		}

		if _, err := l.client.Redis.
			Hset(fmt.Sprintf("%v:%v:%v", common.EmojiKey, data.GuildID.String(), v.ID.String()), v); err != nil {
			log.Fatalf("[%v] Couldn't perform HSET: %v", l.ListenerInfo().Event, err)
		}
	}

	payload, _ := json.Marshal(struct {
		Old []discord.Emoji `json:"old"`
		gateway.EventGuildEmojisUpdate
	}{
		EventGuildEmojisUpdate: data,
		Old:                    emojis,
	})

	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), payload); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}
}

func (l GuildEmojisUpdateListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeGuildEmojisUpdate,
	}
}

func RegisterGuildEmojisUpdateListener(client *lib.GatewayClient) {
	common.RegisterListener(&GuildEmojisUpdateListener{*client})
}
