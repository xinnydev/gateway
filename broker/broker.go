package broker

import (
	log "github.com/disgoorg/log"
	"time"

	"github.com/streadway/amqp"
)

type Broker struct {
	Conn         *amqp.Connection
	Channel      *amqp.Channel
	retryAttempt int
	clientId     string
}

func (b *Broker) Publish(key string, body []byte) error {
	log.Debugf("[amqp] publishing: %v", key)
	return b.Channel.Publish(b.clientId, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Type:        key,
		Body:        body,
	})
}

func NewBroker(clientId, amqpURI string) *Broker {
	b := &Broker{clientId: clientId}
	var err error

	for {
		b.Conn, err = amqp.DialConfig(amqpURI, amqp.Config{
			Dial: amqp.DefaultDial(time.Second * 5),
		})
		if err == nil {
			break
		}
		log.Warn("Failed to connect to AMQP broker. Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}

	b.Channel, err = b.Conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %s", err)
	}

	go b.handleReconnect(amqpURI)

	return b
}

func (b *Broker) handleReconnect(amqpURI string) {
	onClose := b.Conn.NotifyClose(make(chan *amqp.Error))
	for {
		<-onClose
		log.Warnf("AMQP connection lost. Reconnecting...")
		for {
			if b.retryAttempt >= 3 {
				panic("couldn't reconnect to amqp broker after 3 attempts")
			}
			conn, err := amqp.DialConfig(amqpURI, amqp.Config{
				Dial: amqp.DefaultDial(time.Second * 5),
			})
			b.retryAttempt += 1
			if err == nil {
				b.Conn = conn
				b.retryAttempt = 0
				break
			}
			log.Warnf("Failed to connect to AMQP broker. Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		}

		b.Channel, _ = b.Conn.Channel()
		b.handleReconnect(amqpURI)
	}
}
