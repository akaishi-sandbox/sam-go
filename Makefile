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
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o bin/recommend-items ./cmd/recommend-items

local:
	sam local start-api -p 3001 -t ./template.yaml --env-vars ./env.json --region ap-northeast-1

package:
	sam package --region ap-northeast-1 --s3-bucket unisize-artifacts --s3-prefix toc-lambda --template-file ./template.yaml --output-template-file template-packaged.yml

deploy:
	sam deploy --region ap-northeast-1 --template-file template-packaged.yml --stack-name toc-api --capabilities CAPABILITY_NAMED_IAM CAPABILITY_IAM --parameter-overrides $$(jq -r 'to_entries[] | "\(.key)=\(.value)"' ./prod.json)
