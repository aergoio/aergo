# This Makefile is meant to be used by all-in-one build of aergo project

.PHONY: clean protoclean protoc deps test aergosvr aergocli prepare compile build
BINPATH = $(shell pwd)/bin
REPOPATH = github.com/aergoio/aergo
LIBPATH = $(shell pwd)/libtool
LIBTOOLS = luajit

default: compile
	@echo "Done"

prepare: deps

compile: aergocli aergosvr

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

liball:
	@for dir in $(LIBTOOLS); do \
		$(MAKE) PREFIX=$(LIBPATH) -C $(LIBPATH)/src/$$dir all install; \
		if [ $$? != 0 ]; then exit 1; fi; \
	done
	@echo "Done building libs."

test:
	@go test -timeout 60s ./...


clean:
	go clean
	rm -f $(BINPATH)/aergosvr
	rm -f $(BINPATH)/aergocli

protoclean:
	rm -f types/*.pb.go
