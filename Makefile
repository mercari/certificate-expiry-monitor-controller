BINARY = certificate-expiry-monitor-controller
PACKAGES = $(shell go list ./...)

dep:
	@dep ensure -v

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

.PHONY: dep build container push test lint vet coverage
