GOTEST_PACKAGES = $(shell go list ./...)

gomod:
	go mod download

gotest: gomod
	go test -race -v -cover -coverprofile coverage.out $(GOTEST_PACKAGES)

gobench: gomod
	go test -race -bench=. -benchmem $(GOTEST_PACKAGES)

golint: .download_linter
	golangci-lint run -v

.download_linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
