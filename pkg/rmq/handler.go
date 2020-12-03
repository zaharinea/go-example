package rmq

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/service"
)

// Handler struct
type Handler struct {
	config   *config.Config
	services *service.Service
}

// NewHandler returns a new RmqHandler struct
func NewHandler(config *config.Config, services *service.Service) *Handler {
	return &Handler{config: config, services: services}
}

// SetupExchangesAndQueues setup Exchanges and Queues
func SetupExchangesAndQueues(consumer *Consumer, h *Handler) {
	companyQueque := NewQueue("go-example-companies", "events.companies", false, amqp.Table{})
	companyQueque.SetHandler(h.HandlerCompanyEvent)
	companyExchange := NewExchange("events.companies", "fanout", amqp.Table{}, []*Queue{companyQueque})
	consumer.RegisterExchange(companyExchange)

	accountFailedQueque := NewQueue("go-example-accounts-failed", "", false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": "go-example-accounts",
		"x-message-ttl":             60 * 1000,
	})
	accountQueque := NewQueue("go-example-accounts", "", false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": "go-example-accounts-failed",
	})
	accountQueque.SetHandler(h.HandlerAccountEvent)
	consumer.RegisterQueue(accountQueque, accountFailedQueque)
}

// HandlerCompanyEvent handler for company events
func (h *Handler) HandlerCompanyEvent(msg amqp.Delivery) bool {
	if msg.Body == nil {
		logrus.Warning("Error, no message body!")
		return false
	}
	fmt.Println("company event: ", string(msg.Body))
	return true
}

// HandlerAccountEvent handler for account events
func (h *Handler) HandlerAccountEvent(msg amqp.Delivery) bool {
	if msg.Body == nil {
		logrus.Warning("Error, no message body!")
		return false
	}
	fmt.Println("account event: ", string(msg.Body))
	return true
}
