all: test

init:
	go get -u github.com/golang/lint/golint
	go get -u gopkg.in/urfave/cli.v1
	go get -u gopkg.in/xenolog/go-tiny-logger.v1

test:
	go test -v ./...

lint:
	golint ./...

docker-run:
	docker run --rm -it -v "${GOPATH}/src":/go/src/ -w /go/src/github.com/xenolog/liana  golang:alpine  go run liana.go --debug server --password="${LIANA_PASSWORD}" --interfaces=eth0

docker-build:
	docker run --rm -it -v "${GOPATH}/src":/go/src/ -w /go/src/github.com/xenolog/liana  golang:alpine  go build

docker-shell:
	docker run --rm -it -v "${GOPATH}/src":/go/src/ -w /go/src/github.com/xenolog/liana  xenolog/bird  /bin/bash

