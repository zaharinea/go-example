package rmq

import (
	"errors"
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// Queue struct
type Queue struct {
	Name              string
	RoutingKey        string
	Durable           bool
	AutoDelete        bool
	Exclusive         bool
	NoWait            bool
	Arguments         amqp.Table
	requeue           bool
	prefetchCount     int
	handler           func(msg amqp.Delivery) bool
	deliveries        chan amqp.Delivery
	countWorkers      int
	quitWorkerChanels []chan bool
}

// NewQueue returns a new Queue struct
func NewQueue(name string, routingKey string, arguments amqp.Table) *Queue {
	prefetchCount := 2
	multiplier := 1
	countWorkers := prefetchCount * multiplier
	quitWorkerChanels := []chan bool{}
	deliveries := make(chan amqp.Delivery)
	return &Queue{
		Name:              name,
		RoutingKey:        routingKey,
		Durable:           true,
		AutoDelete:        false,
		Exclusive:         false,
		NoWait:            false,
		Arguments:         arguments,
		requeue:           true,
		prefetchCount:     prefetchCount,
		deliveries:        deliveries,
		countWorkers:      countWorkers,
		quitWorkerChanels: quitWorkerChanels,
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
		return fmt.Errorf("Queue Consume: %s", err)
	}

	go func() {
		for delivery := range deliveries {
			q.deliveries <- delivery
		}
	}()

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

// Logger represent common interface for logging function
type Logger interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatal(args ...interface{})
	Infof(format string, args ...interface{})
	Info(args ...interface{})
	Warnf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
}

// Consumer struct
type Consumer struct {
	uri              string
	conn             *amqp.Connection
	channel          *amqp.Channel
	queues           map[string]*Queue
	exchanges        map[string]*Exchange
	err              chan error
	quitReconnector  chan bool
	reconnectTimeout time.Duration
	logger           Logger
}

// NewConsumer returns a new Consumer struct
func NewConsumer(uri string, logger Logger) *Consumer {
	exchanges := make(map[string]*Exchange)
	queues := make(map[string]*Queue)
	err := make(chan error)
	quitReconnector := make(chan bool)
	return &Consumer{
		uri:              uri,
		exchanges:        exchanges,
		queues:           queues,
		err:              err,
		quitReconnector:  quitReconnector,
		reconnectTimeout: time.Second * 3,
		logger:           logger,
	}
}

//Start start consumer
func (c *Consumer) Start() {
	err := c.connect()
	if err != nil {
		c.logger.Fatal("Failed connect", err)
	}

	err = c.setupChanels()
	if err != nil {
		c.logger.Fatal("Failed setup Channel", err)
	}

	err = c.setupQueues()
	if err != nil {
		c.logger.Fatal("Failed setup queues", err)
	}

	err = c.setupExchanges()
	if err != nil {
		c.logger.Fatal("Failed setup exchanges", err)
	}

	err = c.consume()
	if err != nil {
		c.logger.Fatal("Failed setup consumers", err)
	}
}

func (c *Consumer) notifyWorkersQuit() {
	for _, queue := range c.queues {
		if queue.handler == nil {
			continue
		}

		for _, ch := range queue.quitWorkerChanels {
			ch <- true
		}

	}
}

func (c *Consumer) notifyReconnectorQuit() {
	c.quitReconnector <- true
}

//Close stop consumer
func (c *Consumer) Close() error {
	c.notifyReconnectorQuit()
	c.notifyWorkersQuit()

	err := c.channel.Close()
	if err != nil {
		return err
	}
	err = c.conn.Close()
	if err != nil {
		return err
	}
	c.logger.Info("Closing rabbitmq channels and connection")
	return nil
}

//RegisterQueue register queue
func (c *Consumer) RegisterQueue(queue *Queue) {
	if _, exist := c.queues[queue.Name]; exist {
		c.logger.Fatalf("Queue already registred: %s", queue.Name)
	}
	c.queues[queue.Name] = queue
}

//RegisterExchange register exchange
func (c *Consumer) RegisterExchange(exchange *Exchange) {
	for _, queue := range exchange.Queues {
		c.RegisterQueue(queue)
	}

	if _, exist := c.exchanges[exchange.Name]; exist {
		c.logger.Fatalf("Exchange already registred: %s", exchange.Name)
	}
	c.exchanges[exchange.Name] = exchange
}

func (c *Consumer) connect() error {
	var err error
	for {
		c.logger.Info("Start connect to rabbitmq")
		c.conn, err = amqp.Dial(c.uri)
		if err == nil {
			go func() {
				<-c.conn.NotifyClose(make(chan *amqp.Error))
				c.err <- errors.New("Connection Closed")
			}()
			c.logger.Info("Success connect to rabbitmq")
			return nil
		}
		c.logger.Errorf("Failed connect to rabbitmq: %s", err.Error())
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
	c.logger.Debug("Success setup chanels to rabbitmq")
	return nil
}

func (c *Consumer) setupExchanges() error {
	for _, exchange := range c.exchanges {
		if err := exchange.declareAndBind(c.channel); err != nil {
			return err
		}
	}
	c.logger.Debug("Success setup exchanges in rabbitmq")
	return nil
}

func (c *Consumer) setupQueues() error {
	for _, queue := range c.queues {
		if err := queue.declare(c.channel); err != nil {
			return err
		}
	}
	c.logger.Debug("Success setup queues in rabbitmq")
	return nil
}

func (c *Consumer) consume() error {
	c.logger.Debug("Start consume queues")
	for _, queue := range c.queues {
		if queue.handler == nil {
			continue
		}

		if err := queue.consume(c.channel); err != nil {
			return err
		}
		for i := 0; i < queue.countWorkers; i++ {
			quit := make(chan bool)
			queue.quitWorkerChanels = append(queue.quitWorkerChanels, quit)
			go c.consumeHandler(queue, i)
		}
	}

	// watcher for reconnect
	go func() {
		for {
			select {
			case <-c.quitReconnector:
				c.logger.Debugf("Stopped watcher for reconnect")
				return
			case err := <-c.err:
				if err != nil {
					if err := c.reconnect(); err != nil {
						c.logger.Errorf("Failed reconnect to rabbitmq: %s", err)
					}
				}
			}

		}
	}()
	return nil
}

func (c *Consumer) reconsume() error {
	c.logger.Debug("Start reconsume queues")
	for _, queue := range c.queues {
		if queue.handler == nil {
			continue
		}

		if err := queue.consume(c.channel); err != nil {
			return fmt.Errorf("Failed reconsume queue=%s after reconnect: %s", queue.Name, err)
		}
		c.logger.Debugf("Success reconsume queue=%s after reconnect", queue.Name)
	}
	return nil
}

func (c *Consumer) consumeHandler(queue *Queue, workerNumber int) {
	c.logger.Debugf("Start process events: queue=%s, worker=%d", queue.Name, workerNumber)
	for {
		select {
		case delivery := <-queue.deliveries:
			c.logger.Debugf("Got event: queue=%s, worker=%d", queue.Name, workerNumber)
			if queue.handler(delivery) {
				if err := delivery.Ack(false); err != nil {
					c.logger.Errorf("Falied ack %s", queue.Name)
				}
			} else {
				if err := delivery.Nack(false, queue.requeue); err != nil {
					c.logger.Errorf("Falied nack %s", queue.Name)
				}
			}
		case <-queue.quitWorkerChanels[workerNumber]:
			c.logger.Debugf("Stop process events: queue=%s, worker=%d", queue.Name, workerNumber)
			return
		}
	}
}
