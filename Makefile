build:
	go build -v ./...
.PHONY: build

test:
	go test -v ./...
.PHONY: test

install:
	go install -v ./cmd/fdnssearch
.PHONY: install

docker:
	docker build -t nscuro/fdnssearch:latest -f ./build/docker/Dockerfile .
.PHONY: docker
