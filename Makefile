build: tidy
	go build ./...

test:
	go test -v ./... -coverprofile=coverage.out -covermode=atomic -coverpkg=./...

coverage-html:
	go tool cover -html=coverage.out

imports:
	gofumpt -l -w .

lint:
	golangci-lint run --fix ./...

tidy:
	go mod tidy