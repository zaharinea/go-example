package rmq

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/service"
)

// RmqHandler struct
type RmqHandler struct {
	config   *config.Config
	services *service.Service
}

// NewRmqHandler returns a new RmqHandler struct
func NewRmqHandler(config *config.Config, services *service.Service) *RmqHandler {
	return &RmqHandler{config: config, services: services}
}

// SetupExchangesAndQueues setup Exchanges and Queues
func (h *RmqHandler) SetupExchangesAndQueues(consumer *Consumer) {
	companyQueque := NewQueue("go-example-companies", "events.companies", amqp.Table{})
	companyQueque.SetHandler(h.HandlerCompanyEvent)
	companyExchange := NewExchange("events.companies", "fanout", amqp.Table{}, []*Queue{companyQueque})
	consumer.RegisterExchange(companyExchange)

	accountQueque := NewQueue("go-example-accounts", "events.accounts", amqp.Table{})
	accountQueque.SetHandler(h.HandlerAccountEvent)
	consumer.RegisterQueue(accountQueque)
}

// HandlerCompanyEvent handler for company events
func (h *RmqHandler) HandlerCompanyEvent(msg amqp.Delivery) bool {
	if msg.Body == nil {
		logrus.Warning("Error, no message body!")
		return false
	}
	fmt.Println("company event: ", string(msg.Body))
	return true
}

// HandlerAccountEvent handler for account events
func (h *RmqHandler) HandlerAccountEvent(msg amqp.Delivery) bool {
	if msg.Body == nil {
		logrus.Warning("Error, no message body!")
		return false
	}
	fmt.Println("account event: ", string(msg.Body))
	return true
}
