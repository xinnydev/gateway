package common

import (
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
)

var (
	Listeners = map[string]Listener{}
)

type ListenerInfo struct {
	Event gateway.EventType
}

type Listener interface {
	ListenerInfo() *ListenerInfo
	Run(shardId int, ev gateway.EventData)
}

func RegisterListener(listener Listener) {
	Listeners[string(listener.ListenerInfo().Event)] = listener
	log.Infof("registered listener: %v", listener.ListenerInfo().Event)
}
