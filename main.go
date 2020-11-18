package main

import (
	"context"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/handler"
	"github.com/zaharinea/go-example/pkg/repository"
	"github.com/zaharinea/go-example/pkg/service"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Trainer struct
// type Trainer struct {
// 	Name string
// 	Age  int
// 	City string
// }

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

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	logrus.SetLevel(logrus.DebugLevel)

	config := config.NewConfig()
	db := initDb(config)
	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(config, services)

	r := handlers.InitRouter()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	err := r.Run(config.AppAddr)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
