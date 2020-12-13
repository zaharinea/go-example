package rmq

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	rmqclient "github.com/zaharinea/go-rmq-client"
)

func loggingMiddleware(handler rmqclient.HandlerFunc) rmqclient.HandlerFunc {
	return func(ctx context.Context, msg amqp.Delivery) bool {
		queueName := ctx.Value(rmqclient.QueueNameKey).(string)
		logrus.Debugf("start processing event: queue=%s, msg=%s", queueName, string(msg.Body))
		res := handler(ctx, msg)
		logrus.Debugf("end processing event: queue=%s, msg=%s", queueName, string(msg.Body))
		return res
	}
}

var eventCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "rmq_events_total",
	},
	[]string{"queue_name"},
)

var eventProcessedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "rmq_events_processed_total",
	},
	[]string{"queue_name"},
)

var eventFaliedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "rmq_events_falied_total",
	},
	[]string{"queue_name"},
)

var eventProcessingHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "rmq_events_processing_time",
	},
	[]string{"queue_name"},
)

func init() {
	prometheus.MustRegister(eventCounter)
	prometheus.MustRegister(eventProcessedCounter)
	prometheus.MustRegister(eventFaliedCounter)
	prometheus.MustRegister(eventProcessingHistogram)
}

func prometheusMiddleware(handler rmqclient.HandlerFunc) rmqclient.HandlerFunc {
	return func(ctx context.Context, msg amqp.Delivery) bool {
		start := time.Now()
		queueName := ctx.Value(rmqclient.QueueNameKey).(string)

		eventCounter.WithLabelValues(queueName).Inc()
		result := handler(ctx, msg)
		if result == true {
			eventProcessedCounter.WithLabelValues(queueName).Inc()
		} else {
			eventFaliedCounter.WithLabelValues(queueName).Inc()
		}
		duration := time.Since(start)
		eventProcessingHistogram.WithLabelValues(queueName).Observe(duration.Seconds())

		return result
	}
}
