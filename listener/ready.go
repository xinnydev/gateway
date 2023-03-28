package listener

import (
	"encoding/json"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type ReadyListener struct {
	client lib.GatewayClient
}

func (l ReadyListener) Run(ev gateway.EventData) {
	data := ev.(gateway.EventReady)
	log.Infof("%v is ready", data.User.Tag())
	body, _ := json.Marshal(data)
	if err := l.client.Broker.Publish(string(l.ListenerInfo().Event), body); err != nil {
		log.Fatalf("[%v] Couldn't publish exchange: %v", l.ListenerInfo().Event, err)
		return
	}
}

func (l ReadyListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeReady,
	}
}

func RegisterReadyListener(client *lib.GatewayClient) {
	common.RegisterListener(&ReadyListener{*client})
}
