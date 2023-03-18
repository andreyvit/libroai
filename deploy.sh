#!/bin/bash
server=rhodes.tarantsov.com

set -xeuo pipefail
GOOS=linux GOARCH=amd64 go build -o /tmp/buddyd-linux-amd64 ./cmd/buddyd
scp /tmp/buddyd-linux-amd64 "$server:~/"
op inject -i deploy.json | ssh $server 'sudo ~/buddyd-linux-amd64 --install'
