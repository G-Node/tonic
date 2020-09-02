SOURCES = $(shell find . -type f -iname "*.go") go.mod go.sum

.PHONY: clean test showcoverage

all: utonics

utonics: $(SOURCES)
	mkdir -p build
	go build -v -o ./build ./utonics/...

test: $(SOURCES)
	go test -coverpkg=./... -coverprofile=coverage ./...

showcoverage: test
	go tool cover -html=coverage

clean:
	rm -rf coverage build/
