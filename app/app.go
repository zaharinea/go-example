package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/handler"
	"github.com/zaharinea/go-example/pkg/repository"
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
func InitPrometheus(app *gin.Engine) {
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
	p.Use(app)
}

// NewApp return new gin
func NewApp(config *config.Config) *gin.Engine {
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

	app := gin.New()
	app.Use(handler.SetRequestIDMiddleware())
	app.Use(gin.LoggerWithConfig(
		gin.LoggerConfig{
			SkipPaths: []string{"/api/healthcheck", "/metrics"},
			Formatter: func(param gin.LogFormatterParams) string {
				return fmt.Sprintf("%v method: %s path: %s response_time: %v status: %d request_id: %s\n",
					param.TimeStamp.Format(time.RFC3339),
					param.Method,
					param.Path,
					param.Latency.Seconds(),
					param.StatusCode,
					param.Keys[handler.ContextRequestIDKey],
				)
			}}))
	app.Use(gin.Recovery())
	app.Use(sentrygin.New(sentrygin.Options{Repanic: true}))

	InitPrometheus(app)
	handlers.InitRoutes(app)

	return app
}
