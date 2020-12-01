package app

import (
	"strings"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/handler"
	"github.com/zaharinea/go-example/pkg/repository"
	"github.com/zaharinea/go-example/pkg/rmq"
	"github.com/zaharinea/go-example/pkg/service"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

// InitLogger initialize logger
func InitLogger(config *config.Config) {
	if strings.ToUpper(config.LogFormat) == "JSON" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
}

// InitPrometheus initialize prometheus
func InitPrometheus(engine *gin.Engine) {
	p := ginprometheus.NewPrometheus("gin")
	p.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		url := c.Request.URL.Path
		for _, p := range c.Params {
			if p.Key == "id" {
				url = strings.Replace(url, p.Value, ":id", 1)
				break
			}
		}
		return url
	}
	p.Use(engine)
}

// App struct
type App struct {
	Engine      *gin.Engine
	RmqConsumer *rmq.Consumer
}

// NewApp return new gin engine
func NewApp(config *config.Config) *App {
	InitLogger(config)

	err := sentry.Init(sentry.ClientOptions{Dsn: config.SentryDSN, Release: config.AppVersion})
	if err != nil {
		logrus.Errorf("Sentry initialization failed: %v\n", err)
	}

	dbClient := repository.InitDbClient(config)
	repository.ApplyDbMigrations(config, dbClient)
	repos := repository.NewRepository(dbClient.Database(config.MongoDbName))
	services := service.NewService(repos)
	handlers := handler.NewHandler(config, services)

	rmqConsumer := rmq.NewConsumer(config.RmqURI)
	rmqHandlers := rmq.NewRmqHandler(config, services)
	rmqHandlers.SetupExchangesAndQueues(rmqConsumer)

	engine := gin.New()
	engine.Use(handler.SetRequestIDMiddleware())
	engine.Use(handler.Logging())
	engine.Use(handler.Recovery(handler.RecoveryHandler))
	engine.Use(sentrygin.New(sentrygin.Options{Repanic: true}))

	InitPrometheus(engine)
	handlers.InitRoutes(engine)

	return &App{Engine: engine, RmqConsumer: rmqConsumer}
}
