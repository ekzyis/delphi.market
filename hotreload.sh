#!/usr/bin/env bash

PID=$(pidof delphi.market)

set -e

function restart_server() {
  set +e
  [[ -z "$PID" ]] || kill -15 $PID
  ENV=development make build -B
  set -e
  tailwindcss -i public/css/_tw-input.css -o public/css/tailwind.css
  templ generate -path server/router/pages
  go build -o delphi.market .
  ./delphi.market >> server.log 2>&1 &
  templ generate -path server/router/pages
  PID=$(pidof delphi.market)
}

function restart() {
  restart_server
  date +%s.%N > public/hotreload
}

function cleanup() {
    rm -f public/hotreload
    [[ -z "$PID" ]] || kill -15 $PID
}
trap cleanup EXIT

restart
tail -f server.log &

while inotifywait -r -e modify db/ env/ lib/ lnd/ public/ server/; do
  restart
done
