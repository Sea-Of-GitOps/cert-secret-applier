run:
	go run cmd/*.go

build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/main cmd/*.go

test:
	go test ./internal...

testCoverage:
	go test -coverprofile=coverage.out  ./internal...
	go tool cover -html=coverage.out

totalCoverage:
	go test -v -coverpkg=./internal... -coverprofile=profile.cov ./internal...
	go tool cover -func profile.cov | fgrep total | awk '{print substr($$3, 1, length($$3-1))}'

swagger:
	swag init --parseDependency --parseInternal --parseDepth 1 -g ./cmd/*.go

lint:
	golangci-lint run -v

update:
	go get -u -t ./... && go get -u=patch all && go mod tidy