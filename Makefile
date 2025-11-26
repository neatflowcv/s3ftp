.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux go build -o . ./cmd/...

.PHONY: install
install:
	CGO_ENABLED=0 GOOS=linux go install ./cmd/...

.PHONY: update
update:
	go get -u -t ./...
	go mod tidy
	go mod vendor

.PHONY: lint
lint:
	# go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2
	golangci-lint run

.PHONY: fix
fix:
	# go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2
	golangci-lint run --fix

.PHONY: test
test:
	go test -race -shuffle=on ./...

.PHONY: validate
validate: fix test

.PHONY: cover
cover:
	go test ./... --coverpkg ./... -coverprofile=c.out
	go tool cover -html="c.out"
	rm c.out