.PHONY: build run test

SOURCE := $(shell find db env lib lnd pages public server -type f)

build: delphi.market

delphi.market: $(SOURCE)
	go build -o delphi.market .

run:
	go run .

test:
	go test -v -count=1 ./server/router/handler/...
