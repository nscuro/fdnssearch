build:
	go build -v ./...
.PHONY: build

test:
	go test -v ./...
.PHONY: test

install:
	go install -v ./cmd/fdnssearch
.PHONY: install
