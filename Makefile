.PHONY: build run test

SOURCE := $(shell find db env lib lnd pages public server -type f) main.go

build: delphi.market

delphi.market: $(SOURCE)
	templ generate -path server/router/pages
	go build -o delphi.market .

run:
	templ generate -path server/router/pages
	go run .

test:
	go test -v -count=1 ./server/router/handler/...
