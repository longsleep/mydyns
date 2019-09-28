FROM golang:1.13-buster

WORKDIR /go/src/github.com/longsleep/mydyns
COPY . .
RUN make binarystatic

FROM alpine:3.10.2
RUN apk --no-cache add bind-tools

WORKDIR /app
COPY --from=0 /go/src/github.com/longsleep/mydyns/bin/mydynsd.static /app/mydynsd

EXPOSE 38040

VOLUME ["/data"]

ENTRYPOINT ["/app/mydynsd"]
