package main

import (
	"net/http"
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

	s := &http.Server{
		Addr:           c.AppAddr,
		Handler:        a,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
	logrus.Fatal(s.ListenAndServe())
}
