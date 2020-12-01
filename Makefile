ifneq (,$(wildcard .env))
    $(info Found .env file.)
    include .env
	export
endif

SERVICE_NAME=go-example
BINARY_NAME=go-example
PACKAGES ?= $(shell go list -mod=mod ./... | grep -v /vendor)
GOPATH = $(shell go env GOPATH)
DOCKER_COMPOSE=docker-compose -f docker-compose.yml
DOCKER_COMPOSE_TEST=docker-compose -f docker-compose.yml -f docker-compose.test.yml

install-tools:
	go get golang.org/x/lint/golint
	go get github.com/kisielk/errcheck

swagger:
	$(GOPATH)/bin/swag init -g cmd/server/main.go

build: swagger
	go build -o $(BINARY_NAME) cmd/server/main.go

run: build
	./$(BINARY_NAME)

clean:
	go clean

lint: install-tools
	go vet -mod=mod $(PACKAGES)
	$(GOPATH)/bin/golint $(PACKAGES)
	$(GOPATH)/bin/errcheck $(PACKAGES)

test:
	go test -v -cover ./...

new-migration:
	migrate create -ext mongodb -dir migrations -seq $(NAME)

apply-migrations:
	migrate -path migrations -database ${MONGODB_CONNECTION_STRING}/${MONGO_DBNAME} -verbose up

revert-migrations:
	migrate -path migrations -database ${MONGODB_CONNECTION_STRING}/${MONGO_DBNAME} -verbose down


docker-build: swagger
	$(DOCKER_COMPOSE) build

docker-run: docker-build
	$(DOCKER_COMPOSE) up

docker-test:
	$(DOCKER_COMPOSE_TEST) build
	$(DOCKER_COMPOSE_TEST) run --rm $(SERVICE_NAME) go test -v -cover ./...
