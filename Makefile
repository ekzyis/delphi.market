.PHONY: build run test

SOURCE := $(shell find db env lib lnd pages public server -type f) main.go

build: delphi.market

delphi.market: $(SOURCE)
	tailwindcss -i public/css/_tw-input.css -o public/css/tailwind.css
	templ generate -path server/router/pages
	go build -o delphi.market .

run:
	tailwindcss -i public/css/_tw-input.css -o public/css/tailwind.css
	templ generate -path server/router/pages
	go run .

dev:
	bash hotreload.sh

test:
	go test -v -count=1 ./server/router/handler/...
