package destination

import (
	"fmt"
	"github.com/dmotylev/goproperties"
	"github.com/fimston/connection-string-builder"
	"github.com/streadway/amqp"
	"log"
	"time"
)

const AmqpReconnectionInterval = "60s"

type OnDisconnected func()

type RabbitMq struct {
	rabbitMqConnection *amqp.Connection
	rabbitMqChannel    *amqp.Channel
	connString         string
	reconnectDelay     time.Duration
	exchange           string
	routing_key        string
	stopCh             chan bool
}

func onDisconnected(inst *RabbitMq) error {
	log.Printf("[RabbitMq] Connection closed - reconnecting after %s", inst.reconnectDelay)
	for {
		time.Sleep(inst.reconnectDelay)
		if err := inst.dial(); err == nil {
			break
		}
		log.Printf("[RabbitMq] Failed to reconnect. Retry after %s", inst.reconnectDelay)
	}
	log.Printf("[RabbitMq] Connection restored")
	return nil

}

func (self *RabbitMq) dial() error {
	var err error = nil
	self.rabbitMqConnection, err = amqp.Dial(self.connString)
	if err != nil {
		return err
	}
	self.rabbitMqChannel, err = self.rabbitMqConnection.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}

	go func() {
		chClosed := self.rabbitMqChannel.NotifyClose(make(chan *amqp.Error))
		<-chClosed
		onDisconnected(self)
	}()

	return nil
}

func NewRabbitMq(config *properties.Properties) (*RabbitMq, error) {
	connBuilder, err := connstring.CreateBuilder(connstring.ConnectionStringAmqp)
	connBuilder.Address(config.String("rabbitmq.es.addr", ""))
	connBuilder.Port(uint16(config.Int("rabbitmq.es.port", 5672)))
	connBuilder.Username(config.String("rabbitmq.es.username", ""))
	connBuilder.Password(config.String("rabbitmq.es.password", ""))

	reconnectDelay, err := time.ParseDuration(config.String("rabbitmq.es.reconnection_interval", AmqpReconnectionInterval))
	if err != nil {
		return nil, err
	}
	instance := &RabbitMq{
		connString:     connBuilder.Build(),
		reconnectDelay: reconnectDelay,
		exchange:       config.String("rabbitmq.es.exchange.name", ""),
		routing_key:    config.String("rabbitmq.es.routing_key", ""),
	}
	err = instance.dial()
	if err != nil {
		instance = nil
		return nil, err
	}
	return instance, nil
}

func (self *RabbitMq) Push(data []byte) error {
	if err := self.rabbitMqChannel.Publish(
		self.exchange,
		self.routing_key,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            data,
			DeliveryMode:    amqp.Persistent, // 1=non-persistent, 2=persistent
			Priority:        0,               // 0-9
		},
	); err != nil {
		return err
	}
	return nil
}

func (self *RabbitMq) Stop() {
	if self.rabbitMqConnection != nil {
		self.rabbitMqConnection.Close()
	}
}
