package listener

import (
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
}

func (l ReadyListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeReady,
	}
}

func RegisterReadyListener(client *lib.GatewayClient) {
	common.RegisterListener(&ReadyListener{*client})
}
