package rmq

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type Queue struct {
	Name          string
	RoutingKey    string
	Durable       bool
	AutoDelete    bool
	Exclusive     bool
	NoWait        bool
	Arguments     amqp.Table
	requeue       bool
	prefetchCount int
	handlers      []func(msg amqp.Delivery) bool
}

func NewQueue(name string, routingKey string, arguments amqp.Table) *Queue {
	handlers := []func(msg amqp.Delivery) bool{}
	return &Queue{
		Name:          name,
		RoutingKey:    routingKey,
		Durable:       true,
		AutoDelete:    false,
		Exclusive:     false,
		NoWait:        false,
		Arguments:     arguments,
		requeue:       true,
		prefetchCount: 1,
		handlers:      handlers,
	}
}

func (q *Queue) RegisterHandler(handler func(msg amqp.Delivery) bool) {
	q.handlers = append(q.handlers, handler)
}

func (q *Queue) declare(channel *amqp.Channel) error {
	_, err := channel.QueueDeclare(q.Name, q.Durable, q.AutoDelete, q.Exclusive, q.NoWait, q.Arguments)
	if err != nil {
		return fmt.Errorf("Failed to declare a queue %s: %s", q.Name, err)
	}
	return nil
}

func (q *Queue) consume(channel *amqp.Channel) (<-chan amqp.Delivery, error) {
	err := channel.Qos(q.prefetchCount, 0, false)
	if err != nil {
		return nil, fmt.Errorf("Error setting qos: %s", err)
	}

	deliveries, err := channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}
	return deliveries, nil
}

type Exchange struct {
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Arguments  amqp.Table
	Queues     []*Queue
}

func NewExchange(name string, exchangeType string, arguments amqp.Table, queues []*Queue) *Exchange {
	return &Exchange{
		Name:       name,
		Type:       exchangeType,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Arguments:  arguments,
		Queues:     queues,
	}
}

func (e *Exchange) declareAndBind(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(e.Name, e.Type, e.Durable, e.AutoDelete, e.Internal, e.NoWait, e.Arguments); err != nil {
		return fmt.Errorf("Failed to declare an exchange %s: %s", e.Name, err)
	}

	for _, queue := range e.Queues {
		if err := channel.QueueBind(queue.Name, queue.RoutingKey, e.Name, false, nil); err != nil {
			return fmt.Errorf("Failed to bind a queue %s to exchange %s with routing key %s: %s", queue.Name, e.Name, queue.RoutingKey, err)
		}
	}
	return nil
}

type Consumer struct {
	uri              string
	conn             *amqp.Connection
	channel          *amqp.Channel
	queues           map[string]*Queue
	exchanges        map[string]*Exchange
	deliveries       map[string]<-chan amqp.Delivery
	err              chan error
	reconnectTimeout time.Duration
}

func NewConsumer(uri string) *Consumer {
	exchanges := make(map[string]*Exchange)
	queues := make(map[string]*Queue)
	deliveries := make(map[string]<-chan amqp.Delivery)
	err := make(chan error)
	return &Consumer{uri: uri, exchanges: exchanges, queues: queues, deliveries: deliveries, err: err, reconnectTimeout: time.Second * 3}
}

func (c *Consumer) Start() {
	err := c.connect()
	if err != nil {
		logrus.Fatal("Failed connect", err)
	}

	err = c.setupChanels()
	if err != nil {
		logrus.Fatal("Failed setup Channel", err)
	}

	err = c.setupQueues()
	if err != nil {
		logrus.Fatal("Failed setup queues", err)
	}

	err = c.setupExchanges()
	if err != nil {
		logrus.Fatal("Failed setup exchanges", err)
	}

	err = c.setupConsumers()
	if err != nil {
		logrus.Fatal("Failed setup consumers", err)
	}
}

func (c *Consumer) RegisterQueue(queue *Queue) {
	c.queues[queue.Name] = queue
}

func (c *Consumer) RegisterExchange(exchange *Exchange) {
	for _, queue := range exchange.Queues {
		c.RegisterQueue(queue)
	}
	c.exchanges[exchange.Name] = exchange
}

func (c *Consumer) connect() error {
	var err error
	for {
		c.conn, err = amqp.Dial(c.uri)
		if err == nil {
			go func() {
				<-c.conn.NotifyClose(make(chan *amqp.Error))
				c.err <- errors.New("Connection Closed")
			}()
			return nil
		}
		logrus.Errorf("Failed connect to rabbitmq: %s", err.Error())
		time.Sleep(c.reconnectTimeout)

	}
}

func (c *Consumer) reconnect() error {
	if err := c.connect(); err != nil {
		return err
	}
	if err := c.setupChanels(); err != nil {
		return err
	}
	if err := c.setupQueues(); err != nil {
		return err
	}
	return nil
}

func (c *Consumer) setupChanels() error {
	var err error
	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}
	return nil
}

func (c *Consumer) setupExchanges() error {
	for _, exchange := range c.exchanges {
		if err := exchange.declareAndBind(c.channel); err != nil {
			return err
		}
	}
	return nil
}

func (c *Consumer) setupQueues() error {
	for _, queue := range c.queues {
		if err := queue.declare(c.channel); err != nil {
			return err
		}
	}
	return nil
}

func (c *Consumer) setupConsumers() error {
	for _, queue := range c.queues {
		c.consume(queue)
	}
	return nil
}

func (c *Consumer) consume(queue *Queue) error {
	for _, handler := range queue.handlers {
		deliveries, err := queue.consume(c.channel)
		if err != nil {
			return err
		}
		go func() {
			for {
				go func() {
					for delivery := range deliveries {
						if handler(delivery) {
							delivery.Ack(false)
						} else {
							delivery.Nack(false, queue.requeue)
						}
					}
					logrus.Errorf("Rabbit consumer closed: queue=%s", queue.Name)

				}()

				if err := <-c.err; err != nil {
					c.reconnect()
					deliveries, err = queue.consume(c.channel)
					if err != nil {
						logrus.Errorf("Failed consume after reconnect: queue=%s", queue.Name)
					}
				}
			}
		}()
	}
	return nil
}
