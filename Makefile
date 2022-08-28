BINARY = certificate-expiry-monitor-controller
PACKAGES = $(shell go list ./...)

build:
	@go build -o $(BINARY)

test:
	@go test -v -parallel=4 $(PACKAGES)

lint:
	@golint $(PACKAGES)

vet:
	@go vet $(PACKAGES)

coverage:
	@go test -v -race -cover -covermode=atomic -coverprofile=coverage.txt $(PACKAGES)

.PHONY: build container push test lint vet coverage
