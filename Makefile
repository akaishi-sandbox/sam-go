.PHONY: deps clean build

deps:
	go get -u ./...

clean:
	rm -rf ./hello-world/hello-world

build:
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o hello-world/hello-world ./hello-world/main.go
