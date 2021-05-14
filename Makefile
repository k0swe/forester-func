all: build

test:
	go test ./...
	go vet ./...

build:	test
	go build ./...
	cd cmd/forester-func-dev && go build

deploy: build
	git tag deploy -m deploy -s --force && git push --tags --force

clean:
	go clean ./...

.PHONY: test build deploy clean
