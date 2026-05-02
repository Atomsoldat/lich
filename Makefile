.PHONY: format test cover install

format:			## Genericise go code
	@gofmt -s -w .
test:			## Run all tests
	@go clean --testcache && go test ./...
cover:			## Check test coverage
	@go test ./... --coverprofile=cov.out
	@go tool cover --func=cov.out
	@go tool cover --html=cov.out
build:			## Build binary
	@go build
install:		## Install binary
	@sudo install ./lich /usr/local/bin/lich
