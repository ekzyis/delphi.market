.PHONY: build run test dev

SOURCE := $(shell find db env lib lnd public server -type f) main.go


delphi.market: $(SOURCE)
	npm run build
	tailwindcss -i public/css/tw-input.css -o public/css/tailwind.css
	templ generate -path server/router/pages
	go build -o delphi.market .

build: delphi.market

run:
	npm run build
	tailwindcss -i public/css/tw-input.css -o public/css/tailwind.css
	templ generate -path server/router/pages
	go run .

dev:
	bash hotreload.sh

test:
	go test -v -count=1 ./server/router/handler/...
