.DEFAULT_GOAL := build

PROJECT_URL := github.com/MawKKe/integer-interval-expressions-go

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

tmp:
	mkdir -p $@

tmp/coverage.data: | tmp
	go test -coverprofile=$@

.PHONY: tmp/coverage.data

tmp/coverage.html: tmp/coverage.data
	go tool cover -html=$< -o $@

coverage: tmp/coverage.html
	@echo "---"
	@echo "Open $^ in browser to view coverage info"
	@echo "---"

clean:
	go clean -x ./...
	rm -rf tmp

git_latest_version_tag := git describe --tags --match "v[0-9]*" --abbrev=0

# Make sure the tags are published and pushed to the public remote!
sync-package-proxy:
	GOPROXY=proxy.golang.org go list -m ${PROJECT_URL}@$(shell ${git_latest_version_tag})

.PHONY: build test fmt vet clean coverage sync-package-proxy
