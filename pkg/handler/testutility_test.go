package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

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

func performRequest(r http.Handler, method, path string, body string) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, _ := http.NewRequest(method, path, bodyReader)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
