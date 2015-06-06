FROM gliderlabs/alpine:3.1

RUN mkdir /app
ADD bin/mydynsd.static /app/mydynsd

RUN apk --update add bind-tools

EXPOSE 38040

VOLUME ["/data"]

ENTRYPOINT ["/app/mydynsd"]
