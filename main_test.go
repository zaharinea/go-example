package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zaharinea/go-example/config"
	"github.com/zaharinea/go-example/pkg/repository"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	c := config.NewTestingConfig()
	dbClient := repository.InitDbClient(c)
	repository.ApplyDbMigrations(c, dbClient)

	fmt.Printf("\033[1;36m%s\033[0m", "> Setup completed\n")
}

func teardown() {
	// Do something here.
	fmt.Printf("\033[1;36m%s\033[0m", "> Teardown completed\n")
}

//TestOk need for run setup and teardown
func TestOk(t *testing.T) {
	assert.Equal(t, true, true)
}
