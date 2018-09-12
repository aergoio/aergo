# This Makefile is meant to be used by all-in-one build of aergo project

.PHONY: clean protoclean protoc deps test aergosvr aergocli prepare compile build \
	liball liball-clean
BINPATH = $(shell pwd)/bin
REPOPATH = github.com/aergoio/aergo
LIBPATH = $(shell pwd)/libtool

default: compile
	@echo "Done"

prepare: deps

compile: aergocli aergosvr aergoluac aergoscc

build: test compile

all: clean prepare build
	@echo "Done All"


deps: liball
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

aergoluac: ./cmd/aergoluac/*.go
	go build -o $(BINPATH)/aergoluac ./cmd/aergoluac
	@echo "Done buidling aergoluac."

aergoscc: ./cmd/aergoscc/*.go
	go build -o $(BINPATH)/aergoscc ./cmd/aergoscc
	@echo "Done buidling aergoscc."

test:
	@go test -timeout 60s ./...

liball:
	@cd $(LIBPATH) && $(MAKE) install
	@echo "Done installing tools."

liball-clean:
	@cd $(LIBPATH) && $(MAKE) uninstall
	@echo "Done uninstalling tools."

clean: liball-clean
	go clean
	rm -f $(BINPATH)/aergosvr
	rm -f $(BINPATH)/aergocli

protoclean:
	rm -f types/*.pb.go
