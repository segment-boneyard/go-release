build:
	@mkdir -p build
	@go build -o build/release

.PHONY: build build-all
