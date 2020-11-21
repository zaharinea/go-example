package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

// InitDbClient initialize DB Client
func InitDbClient(config *config.Config) *mongo.Client {
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
	return client
}

// ApplyMigrations apply all migrations
func ApplyMigrations(config *config.Config, client *mongo.Client) {
	mConfig := mongodb.Config{
		DatabaseName: config.MongoDbName,
		Locking:      mongodb.Locking{Enabled: true},
	}
	driver, err := mongodb.WithInstance(client, &mConfig)
	if err != nil {
		logrus.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(config.MongoMigrationsDir, "mongodb", driver)
	if err != nil {
		logrus.Fatal(err)
	}
	err = m.Up()
	if err != nil {
		logrus.Fatal(err)
	}
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

// @title Go-example API
// @version 1.0
// @description This is a simple http server
func main() {
	config := config.NewConfig()

	InitLogger(config)

	dbClient := InitDbClient(config)
	ApplyMigrations(config, dbClient)
	repos := repository.NewRepository(dbClient.Database(config.MongoDbName))
	services := service.NewService(repos)
	handlers := handler.NewHandler(config, services)

	app := gin.New()
	InitPrometheus(app)
	handlers.InitRoutes(app)

	app.Use(gin.Logger(), gin.Recovery())

	err := app.Run(config.AppAddr)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
