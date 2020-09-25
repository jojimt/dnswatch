#!/bin/bash
set -e
docker run --rm -m 16g -v ${PWD}:/go/src/github.com/jojimt/dnswatch -w /go/src/github.com/jojimt/dnswatch/cmd/dnswatch -e CGO_ENABLED=0 -e GOOS=linux --network=host -it noirolabs/gobuild1.14 go build -v .

docker build -t jojimt/dnswatch -f Dockerfile .
docker push jojimt/dnswatch
