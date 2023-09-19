build: tidy
	go build ./...

test: build
	go test -v ./...

imports:
	gofumpt -l -w .

lint:
	golangci-lint run --fix ./...

tidy:
	go mod tidy