GO_TEST_PACKAGES = $(shell go list ./...)

gotest:
	go test -race -v -cover -coverprofile coverage.out $(GO_TEST_PACKAGES)

golint:
	golangci-lint run -v

download_linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
