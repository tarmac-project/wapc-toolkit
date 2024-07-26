build:
	$(MAKE) -C testdata/hello-go build

tests: build
	$(MAKE) -C callbacks tests
	$(MAKE) -C engine tests

benchmarks: build
	$(MAKE) -C callbacks benchmarks
	$(MAKE) -C engine benchmarks
