# This Makefile is meant to be used by all-in-one build of aergo project

.PHONY: clean protoclean protoc deps test aergosvr aergocli prepare compile build
BINPATH = $(shell pwd)/bin
REPOPATH = github.com/aergoio/aergo


default: compile
	@echo "Done"

prepare: deps

compile: aergocli aergosvr

build: test compile

all: clean prepare build
	@echo "Done All"


deps:
	glide install

# FIXME: make recursive to subdirectories
protoc:
	protoc -I/usr/local/include \
		-I${GOPATH}/src/${REPOPATH}/types \
		--go_out=plugins=grpc:${GOPATH}/src \
		${GOPATH}/src/${REPOPATH}/types/*.proto 
	go build ./types/...

aergosvr: cmd/aergosvr/*.go
	go build -o $(BINPATH)/aergosvr ./cmd/aergosvr
	@echo "Done buidling aergosvr."

aergocli: cmd/aergocli/*.go
	go build -o $(BINPATH)/aergocli ./cmd/aergocli
	@echo "Done buidling aergocli."

test:
	@go test -timeout 60s ./...


clean:
	go clean
	rm -f $(BINPATH)/aergosvr
	rm -f $(BINPATH)/aergocli

protoclean:
	rm -f types/*.pb.go
