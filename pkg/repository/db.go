package repository

import (
	"context"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
	"github.com/zaharinea/go-example/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

// ApplyDbMigrations apply all migrations
func ApplyDbMigrations(config *config.Config, client *mongo.Client) {
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
		if err == migrate.ErrNoChange {
			return
		}

		logrus.Fatal(err)
	}
}
