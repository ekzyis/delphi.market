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

function sync() {
  restart_server
  date +%s.%N > public/hotreload
  rsync -avh public/ dev1.delphi.market:/var/www/dev1.delphi --delete
}

function cleanup() {
    rm -f public/hotreload
    [[ -z "$PID" ]] || kill -15 $PID
}
trap cleanup EXIT

sync
tail -f server.log &

while inotifywait -r -e modify src/ pages/; do
  sync
done
