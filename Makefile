all: test

init:
	go get -u github.com/golang/lint/golint
	go get -u gopkg.in/urfave/cli.v1
	go get -u gopkg.in/xenolog/go-tiny-logger.v1

test:
	go test -v ./...

lint:
	golint ./...

