.PHONY: all test

all: test

test:
	go test ./... -v -race -coverprofile=test/coverage.txt -covermode=atomic
