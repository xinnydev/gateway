package listener

import (
	"github.com/disgoorg/disgo/gateway"
	"github.com/xinny/gateway/common"
	"github.com/xinny/gateway/lib"
)

type RawListener struct {
	client lib.GatewayClient
}

func (l RawListener) Run(ev gateway.EventData) {
	data := ev.(gateway.EventRaw)
	switch data.EventType {
	// Handle HELLO timeout here
	}
}

func (l RawListener) ListenerInfo() *common.ListenerInfo {
	return &common.ListenerInfo{
		Event: gateway.EventTypeRaw,
	}
}

func RegisterRawListener(client *lib.GatewayClient) {
	common.RegisterListener(&RawListener{*client})
}
