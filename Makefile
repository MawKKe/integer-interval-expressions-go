.DEFAULT_GOAL := build

build:
	go build ./...

test:
	go test ./...

test-verbose:
	go test -v ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

# the fact that this is even needed is idiotic but whatchugonnado #golang #fuckyouweknowbetter
fix:
	find . -type f -iname "*.go" -exec goimports -w {} +

clean:
	go clean -x ./...

.PHONY: build test fmt vet
