package common

import (
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/log"
)

var (
	Listeners []Listener
)

type ListenerInfo struct {
	Event gateway.EventType
}

type Listener interface {
	ListenerInfo() *ListenerInfo
	Run(ev gateway.EventData)
}

func RegisterListener(listener Listener) {
	Listeners = append(Listeners, listener)
	log.Infof("registered listener: %v", listener.ListenerInfo().Event)
}
