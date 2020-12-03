package rmq

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/repository"
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
		logrus.Errorf("Invalid company event: msg=%s", string(msg.Body))
		return false
	}
	logrus.Debugf("company event: msg=%s", string(msg.Body))
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
func (h *Handler) HandlerAccountEvent(msg amqp.Delivery) bool {
	if msg.Body == nil {
		logrus.Errorf("Invalid account event: msg=%s", string(msg.Body))
		return false
	}
	logrus.Debugf("account event: msg=%s", string(msg.Body))

	var account repository.Account
	if err := json.Unmarshal(msg.Body, &account); err != nil {
		logrus.Errorf("Invalid account event: msg=%s, error=%s", string(msg.Body), err)
		return false
	}

	ctx := context.Background()
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
