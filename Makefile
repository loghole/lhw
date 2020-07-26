GO_TEST_PACKAGES = $(shell go list ./...)

go-mod:
	go mod download

go-test: go-mod
	go test -race -v -cover -coverprofile coverage.out $(GO_TEST_PACKAGES)

go-bench: go-mod
	go test -race -bench=. -benchmem $(GO_TEST_PACKAGES)

go-lint:
	golangci-lint run -v

download_linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
