build:
	@mkdir -p build
	@go build -o build/release

release: build
	./build/release segmentio go-release --assets build/release

.PHONY: build
