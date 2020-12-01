package rmq

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// Queue struct
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
	handler       func(msg amqp.Delivery) bool
	deliveries    <-chan amqp.Delivery
	waitReconnect chan bool
}

// NewQueue returns a new Queue struct
func NewQueue(name string, routingKey string, arguments amqp.Table) *Queue {
	waitReconnect := make(chan bool)
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
		waitReconnect: waitReconnect,
	}
}

// SetHandler register handler in Queue
func (q *Queue) SetHandler(handler func(msg amqp.Delivery) bool) {
	q.handler = handler
}

func (q *Queue) declare(channel *amqp.Channel) error {
	_, err := channel.QueueDeclare(q.Name, q.Durable, q.AutoDelete, q.Exclusive, q.NoWait, q.Arguments)
	if err != nil {
		return fmt.Errorf("Failed to declare a queue %s: %s", q.Name, err)
	}
	return nil
}

func (q *Queue) consume(channel *amqp.Channel) error {
	err := channel.Qos(q.prefetchCount, 0, false)
	if err != nil {
		return fmt.Errorf("Error setting qos: %s", err)
	}

	q.deliveries, err = channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("Queue Consume: %s", err)
	}
	return nil
}

// Exchange struct
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

// NewExchange returns a new Exchange struct
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

// Consumer struct
type Consumer struct {
	uri              string
	conn             *amqp.Connection
	channel          *amqp.Channel
	queues           map[string]*Queue
	exchanges        map[string]*Exchange
	err              chan error
	reconnectTimeout time.Duration
}

// NewConsumer returns a new Consumer struct
func NewConsumer(uri string) *Consumer {
	exchanges := make(map[string]*Exchange)
	queues := make(map[string]*Queue)
	err := make(chan error)
	return &Consumer{uri: uri, exchanges: exchanges, queues: queues, err: err, reconnectTimeout: time.Second * 3}
}

//Start start consumer
func (c *Consumer) Start() {
	logrus.Infof("exchanges: %v", c.exchanges)
	logrus.Infof("queues: %v", c.queues)

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

	err = c.consume()
	if err != nil {
		logrus.Fatal("Failed setup consumers", err)
	}
}

//Close stop consumer
func (c *Consumer) Close() error {
	err := c.channel.Close()
	if err != nil {
		return err
	}
	err = c.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

//RegisterQueue register queue
func (c *Consumer) RegisterQueue(queue *Queue) {
	if _, exist := c.queues[queue.Name]; exist {
		logrus.Fatalf("Queue already registred: %s", queue.Name)
	}
	c.queues[queue.Name] = queue
}

//RegisterExchange register exchange
func (c *Consumer) RegisterExchange(exchange *Exchange) {
	for _, queue := range exchange.Queues {
		c.RegisterQueue(queue)
	}

	if _, exist := c.exchanges[exchange.Name]; exist {
		logrus.Fatalf("Exchange already registred: %s", exchange.Name)
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
	if err := c.reconsume(); err != nil {
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

func (c *Consumer) consume() error {
	for _, queue := range c.queues {
		if queue.handler == nil {
			continue
		}

		if err := queue.consume(c.channel); err != nil {
			return err
		}
		go c.consumeHandler(queue)
	}

	// reconnector
	go func() {
		for {
			if err := <-c.err; err != nil {
				logrus.Info("Start reconnect")
				if err := c.reconnect(); err != nil {
					logrus.Errorf("Failed reconnect: %s", err)
				}
			}
			for _, queue := range c.queues {
				queue.waitReconnect <- true
			}
		}
	}()
	return nil
}

func (c *Consumer) reconsume() error {
	for _, queue := range c.queues {
		if queue.handler == nil {
			continue
		}

		if err := queue.consume(c.channel); err != nil {
			return fmt.Errorf("Failed reconsume queue=%s after reconnect: %s", queue.Name, err)
		}
		logrus.Infof("Success reconsume queue=%s after reconnect", queue.Name)
	}
	return nil
}

func (c *Consumer) consumeHandler(queue *Queue) {
	for {
		logrus.Infof("Start process events: queue=%s", queue.Name)

		for delivery := range queue.deliveries {
			if queue.handler(delivery) {
				if err := delivery.Ack(false); err != nil {
					logrus.Errorf("Falied ack %s", queue.Name)
				}
			} else {
				if err := delivery.Nack(false, queue.requeue); err != nil {
					logrus.Errorf("Falied nack %s", queue.Name)
				}
			}
		}
		logrus.Errorf("Rabbit consumer closed: queue=%s", queue.Name)

		<-queue.waitReconnect
	}
}
