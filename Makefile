BINARY_NAME=go-example
PACKAGES ?= $(shell go list -mod=mod ./... | grep -v /vendor)
GOPATH = $(shell go env GOPATH)

MONGODB_CONNECTION_STRING=mongodb://localhost:27017
MONGO_DBNAME=go-example

install:
	go mod vendor

install-tools:
	go get golang.org/x/lint/golint
	go get github.com/kisielk/errcheck

swagger:
	$(GOPATH)/bin/swag init

build: swagger
	go build -o $(BINARY_NAME) main.go

docker-build: swagger
	docker-compose build

run: build
	./$(BINARY_NAME)

docker-run: docker-build
	docker-compose up

clean:
	go clean

lint: install-tools
	go vet -mod=mod $(PACKAGES)
	$(GOPATH)/bin/golint $(PACKAGES)
	$(GOPATH)/bin/errcheck $(PACKAGES)
