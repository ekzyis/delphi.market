build: delphi.market

delphi.market: src/*.go
	go build -o delphi.market ./src/

run:
	go run ./src/

