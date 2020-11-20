package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/handler"
	"github.com/zaharinea/go-example/pkg/repository"
	"github.com/zaharinea/go-example/pkg/service"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// InitDb initialize DB
func InitDb(config *config.Config) *mongo.Database {
	client, err := mongo.NewClient(options.Client().ApplyURI(config.MongoURI))
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		logrus.Error(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		logrus.Error(err)
	}

	return client.Database(config.MongoDB)

}

// InitPrometheus initialize prometheus
func InitPrometheus(r *gin.Engine) {
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
	p.Use(r)
}

// @title Go-example API
// @version 1.0
// @description This is a simple http server
func main() {
	config := config.NewConfig()

	InitLogger(config)

	db := InitDb(config)
	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(config, services)

	r := gin.New()
	InitPrometheus(r)
	handlers.InitRoutes(r)

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	err := r.Run(config.AppAddr)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
