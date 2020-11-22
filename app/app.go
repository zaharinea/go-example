package app

import (
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	dbClient := repository.InitDbClient(config)
	repository.ApplyDbMigrations(config, dbClient)
	repos := repository.NewRepository(dbClient.Database(config.MongoDbName))
	services := service.NewService(repos)
	handlers := handler.NewHandler(config, services)

	app := gin.New()
	InitPrometheus(app)
	handlers.InitRoutes(app)

	app.Use(gin.Logger(), gin.Recovery())
	return app
}
