EXENAME := mydynsd
OUTPUT := $(CURDIR)/bin
GO111MODULE=on

build: binary

binary:
	CGO_ENABLED=0 go build \
		-trimpath \
		-tags release \
		-buildmode=exe \
		-ldflags '-s -w -extldflags -static' \
		-o $(OUTPUT)/$(EXENAME) ./cmd/mydynsd

fmt:
	go fmt ./...

test:
	go test -v ./...

.PHONY: build binary fmt test
