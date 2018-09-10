# This Makefile is meant to be used by all-in-one build of aergo project

.PHONY: build all test clean liball liball-clean deps protoc protoclean 

BINPATH := $(shell pwd)/bin
CMDS := aergocli aergosvr aergoluac
REPOPATH := github.com/aergoio/aergo

build: vendor liball
	GOBIN=$(BINPATH) go install ./cmd/...

all: clean test build
	@echo "Done All"

vendor: glide.yaml glide.lock
	@glide install

# test

test:
	@go test -timeout 60s ./...

# clean

clean: liball-clean
	go clean
	rm -f $(addprefix $(BINPATH)/,$(CMDS))

# 3rd party libs

LIBPATH := $(shell pwd)/libtool

liball: 
	$(MAKE) -C $(LIBPATH) install

liball-clean:
	$(MAKE) -C $(LIBPATH) uninstall

# etc

deps: vendor liball
	@glide install

protoc:
	protoc -I/usr/local/include \
		-I${GOPATH}/src/${REPOPATH}/aergo-protobuf/proto \
		--go_out=plugins=grpc:${GOPATH}/src \
		${GOPATH}/src/${REPOPATH}/aergo-protobuf/proto/*.proto
	go build ./types/...

protoclean:
	rm -f types/*.pb.go

