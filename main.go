package main

import (
	"os"

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

	err := a.Run(c.AppAddr)
	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
}
