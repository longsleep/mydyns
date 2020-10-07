FROM golang:1.15-buster

WORKDIR /go/src/github.com/longsleep/mydyns
COPY . .
RUN make binary

FROM alpine:3.12.0
RUN apk --no-cache add bind-tools

WORKDIR /app
COPY --from=0 /go/src/github.com/longsleep/mydyns/bin/mydynsd /app/mydynsd

EXPOSE 38040

VOLUME ["/data"]

ENTRYPOINT ["/app/mydynsd"]
