SOURCE_FILES?=./...
TEST_PATTERN?=.
TEST_OPTIONS?=


export PATH := ./bin:$(PATH)
export GO111MODULE := on


setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.17.1
	curl -L https://git.io/misspell | sh
.PHONY: setup



# Run all the tests
test:
	go get -u github.com/pierrre/gotestcover
	gotestcover $(TEST_OPTIONS) -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m
.PHONY: cover

# Run all the tests and opens the coverage report
cover: test
	go tool cover -html=coverage.txt
.PHONY: cover

# gofmt and goimports all go files
fmt:
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done
.PHONY: fmt

# Run all the linters
lint:
	golangci-lint --version
	golangci-lint run ./...
	./bin/misspell -error **/*
.PHONY: lint

# Run all the tests and code checks
ci: build test lint
.PHONY: ci

# Build a beta version of assumer
build:
	go build -o assumer cmd/assumer/main.go
.PHONY: build


