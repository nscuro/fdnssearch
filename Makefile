DIST_DIR=./dist
LDFLAGS="-s -w"
GCFLAGS="all=-trimpath=$(shell pwd)"
ASMFLAGS="all=-trimpath=$(shell pwd)"

build:
	go build -v ./...
.PHONY: build

test:
	go test -v ./...
.PHONY: test

bench:
	go test -bench=. ./...
.PHONY: bench

install:
	go install -ldflags=${LDFLAGS} -gcflags=${GCFLAGS} -asmflags=${ASMFLAGS} -v ./cmd/fdnssearch
.PHONY: install

docker:
	docker build -t nscuro/fdnssearch:latest -f ./build/docker/Dockerfile .
.PHONY: docker

pre-dist:
	mkdir -p ${DIST_DIR}
.PHONY: 

windows:
	GOOS=windows GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 \
	go build -ldflags=${LDFLAGS} -gcflags=${GCFLAGS} -asmflags=${ASMFLAGS} \
		-o ${DIST_DIR}/fdnssearch-windows-amd64.exe ./cmd/fdnssearch
.PHONY: windows

darwin:
	GOOS=darwin GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 \
	go build -ldflags=${LDFLAGS} -gcflags=${GCFLAGS} -asmflags=${ASMFLAGS} \
		-o ${DIST_DIR}/fdnssearch-darwin-amd64 ./cmd/fdnssearch	
.PHONY: darwin

linux:
	GOOS=linux GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 \
	go build -ldflags=${LDFLAGS} -gcflags=${GCFLAGS} -asmflags=${ASMFLAGS} \
		-o ${DIST_DIR}/fdnssearch-linux-amd64 ./cmd/fdnssearch
.PHONY: linux

clean:
	rm -rf ${DIST_DIR}; go clean ./...
.PHONY: clean

all: clean build test pre-dist windows darwin linux
.PHONY: all