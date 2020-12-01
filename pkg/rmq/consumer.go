package rmq

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/service"
)

type RmqHandler struct {
	config   *config.Config
	services *service.Service
}

func NewRmqHandler(config *config.Config, services *service.Service) *RmqHandler {
	return &RmqHandler{config: config, services: services}
}

func (h *RmqHandler) SetupExchangesAndQueues(consumer *Consumer) {
	companyQueque := NewQueue("go-example-companies", "events.companies", amqp.Table{})
	companyQueque.RegisterHandler(h.HandlerCompanyEvent)
	companyExchange := NewExchange("events.companies", "fanout", amqp.Table{}, []*Queue{companyQueque})
	consumer.RegisterExchange(companyExchange)
}

func (h *RmqHandler) HandlerCompanyEvent(msg amqp.Delivery) bool {
	if msg.Body == nil {
		logrus.Warning("Error, no message body!")
		return false
	}
	fmt.Println(string(msg.Body))
	return true
}
