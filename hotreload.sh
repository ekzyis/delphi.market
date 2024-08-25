#!/usr/bin/env bash

PID=$(pidof delphi.market)

set -e

echo ":: remote port forwarding for dev1.delphi.market ::"
ssh -fnNR 4322:localhost:4321 dev1.delphi.market
echo

function restart_server() {
  set +e
  [[ -z "$PID" ]] || kill -15 $PID
  ENV=development make build -B
  set -e
  ./delphi.market 2>&1 &
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

while inotifywait -r -e modify db/ env/ lib/ lnd/ public/ server/; do
  restart
done
