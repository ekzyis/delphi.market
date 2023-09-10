#!/usr/bin/env bash

PID=$(pidof delphi.market)

set -e

function restart_server() {
  set +e
  [[ -z "$PID" ]] || kill -15 $PID
  ENV=development make build -B
  set -e
  ./delphi.market >> server.log 2>&1 &
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

while inotifywait -r -e modify db/ env/ lib/ lnd/ pages/ public/ server/; do
  restart
done
