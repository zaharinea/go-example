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

// @title Go-example API
// @version 1.0
// @description This is a simple http server
func main() {
	c := config.NewConfig()
	a := app.NewApp(c)

	a.RmqConsumer.Start()

	srv := &http.Server{
		Addr:           c.AppAddr,
		Handler:        a.Engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logrus.Infof("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutdown Server ...")

	if err := a.RmqConsumer.Close(); err != nil {
		logrus.Infof("Close rabbitmq consumer: %s\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal("Server Shutdown:", err)
	}
	logrus.Info("Server exiting")
}
