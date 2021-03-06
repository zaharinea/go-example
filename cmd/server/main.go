package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zaharinea/go-example/app"
	"github.com/zaharinea/go-example/config"
)

func main() {
	c := config.NewConfig()
	a := app.NewApp(c)

	a.RmqConsumer.Start()

	httpSrv := &http.Server{
		Addr:           c.AppAddr,
		Handler:        a.Engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			logrus.Infof("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutdown Server ...")

	if err := a.RmqConsumer.Stop(); err != nil {
		logrus.Infof("Stop rabbitmq consumer: %s\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		logrus.Fatal("Http server shutdown:", err)
	}

	if err := a.DbClient.Disconnect(ctx); err != nil {
		logrus.Fatal("MongoDB client disconnect:", err)
	}
	logrus.Info("Connection to MongoDB closed")

	logrus.Info("Server exiting")
}
