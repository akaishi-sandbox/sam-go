.PHONY: deps clean build

deps:
	go get -u ./...

clean:
	rm -rf ./bin

test:
	go test  -v ./...

build:
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o bin/api ./
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o bin/search-items ./function/search-items
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o bin/classification-info ./function/classification-info
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o bin/recommend-items ./function/recommend-items
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o bin/access-info ./function/access-info

local:
	sam local start-api -p 3001 -t ./template.yaml --env-vars ./env.json --region ap-northeast-1

package:
	sam package --region ap-northeast-1 --s3-bucket unisize-artifacts-develop --s3-prefix toc-lambda --template-file ./template.yaml --output-template-file packaged.yml

deploy:
	sam deploy --region ap-northeast-1 --template-file packaged.yml --stack-name toc-api --capabilities CAPABILITY_NAMED_IAM CAPABILITY_IAM --parameter-overrides $$(jq -r 'to_entries[] | "\(.key)=\(.value)"' ./prod.json)
