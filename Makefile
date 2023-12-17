build:
	$(MAKE) -C testdata/hello-go build

tests: build
	go test --race -v -covermode=atomic -coverprofile=coverage.out ./...
