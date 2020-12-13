package rmq

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/repository"
	rmqclient "github.com/zaharinea/go-rmq-client"
)

// Handler struct
type Handler struct {
	config *config.Config
	repos  *repository.Repository
}

// NewHandler returns a new RmqHandler struct
func NewHandler(config *config.Config, repos *repository.Repository) *Handler {
	return &Handler{config: config, repos: repos}
}

// SetupExchangesAndQueues setup Exchanges and Queues
func SetupExchangesAndQueues(consumer *rmqclient.Consumer, h *Handler) {
	companyQueque := rmqclient.NewQueue("go-example-companies", "events.companies", amqp.Table{})
	companyQueque.SetHandler(h.HandlerCompanyEvent)
	companyExchange := rmqclient.NewExchange("events.companies", "fanout", amqp.Table{}, []*rmqclient.Queue{companyQueque})
	consumer.RegisterExchange(companyExchange)

	accountFailedQueque := rmqclient.NewQueue("go-example-accounts-failed", "", amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": "go-example-accounts",
		"x-message-ttl":             60 * 1000,
	})
	accountQueque := rmqclient.NewQueue("go-example-accounts", "", amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": "go-example-accounts-failed",
	})
	accountQueque.SetHandler(h.HandlerAccountEvent)
	consumer.RegisterQueue(accountQueque, accountFailedQueque)

	consumer.RegisterMiddleware(loggingMiddleware, prometheusMiddleware)
}

// HandlerCompanyEvent handler for company events
func (h *Handler) HandlerCompanyEvent(ctx context.Context, msg amqp.Delivery) bool {
	if msg.Body == nil {
		logrus.Errorf("Invalid company event: msg=%s", string(msg.Body))
		return false
	}
	return true
}

// HandlerAccountEvent handler for account events
/*
{
    "external_id":"1",
    "name":"account1",
    "created_at":"2020-11-20T22:56:57.565Z",
    "updated_at":"2020-11-20T22:56:57.565Z"
}
*/
func (h *Handler) HandlerAccountEvent(ctx context.Context, msg amqp.Delivery) bool {
	if msg.Body == nil {
		logrus.Errorf("Invalid account event: msg=%s", string(msg.Body))
		return false
	}

	var account repository.Account
	if err := json.Unmarshal(msg.Body, &account); err != nil {
		logrus.Errorf("Invalid account event: msg=%s, error=%s", string(msg.Body), err)
		return false
	}

	if _, err := h.repos.Account.CreateOrUpdate(ctx, account, false); err != nil {
		if h.repos.IsDuplicateKeyErr(err) {
			logrus.Infof("Skip duplicate or expired event: msg=%s", string(msg.Body))
			return true
		}

		logrus.Errorf("Failed create or update account: msg=%s, error=%s", string(msg.Body), err)
		return false
	}

	return true
}
