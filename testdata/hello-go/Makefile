build:
	docker run -v `pwd`/:/build -w /build tinygo/tinygo:0.25.0 tinygo build -o /build/hello.wasm -target wasi /build/main.go
