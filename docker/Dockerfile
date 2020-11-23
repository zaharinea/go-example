FROM golang:1.15-alpine AS build_base

ENV GO111MODULE=on

WORKDIR /tmp/app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .


FROM alpine:3.12
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build_base /tmp/app/main /app/main
COPY --from=build_base /tmp/app/migrations /app/migrations
WORKDIR /app
EXPOSE 8000
CMD ["/app/main"]