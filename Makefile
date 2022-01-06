all: build

test-go:
	go test ./...
	go vet ./...

build-go:	test-go
	go build ./...
	cd cmd/forester-func-dev && go build

build-js:
	cd javascript/functions && npm install && npm run build

deploy: build-go build-js
	git tag deploy -m deploy -s --force && git push --tags --force

clean:
	go clean ./...

.PHONY: test-go build-go build-js deploy clean
