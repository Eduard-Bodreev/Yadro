.PHONY: build build-server run run-server

build:
	go build -o ./cmd/bin/xkcd ./cmd/xkcd

build-server:
	go build -o ./cmd/bin/xkcd-server ./cmd/xkcd-server

run: build
	./cmd/bin/xkcd -c ./config/config.yaml

run-server: build-server
	./cmd/bin/xkcd-server -config ./config/config.yaml -p 8080