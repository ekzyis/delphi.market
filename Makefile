.PHONY: build run

SOURCE := $(shell find db env lib lnd pages public server -type f)

build: delphi.market

delphi.market: $(SOURCE)
	go build -o delphi.market .

run:
	go run .

