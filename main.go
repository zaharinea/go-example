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

func initDb(config *config.Config) *mongo.Database {
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

// InitPrometheus initialize rometheus
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

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	logrus.SetLevel(logrus.DebugLevel)

	config := config.NewConfig()
	db := initDb(config)
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
