FROM alpine:3.10.2
RUN apk upgrade --no-cache
COPY cmd/dnswatch/dnswatch /usr/local/bin/
ENTRYPOINT exec /usr/local/bin/dnswatch
