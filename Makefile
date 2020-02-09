.PHONY: deps clean build

deps:
	go get -u ./...

clean:
	rm -rf ./bin

test:
	go test  -v ./...

build:
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o bin/search-items ./cmd/search-items
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o bin/classification-info ./cmd/classification-info
