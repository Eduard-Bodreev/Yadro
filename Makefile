BINARY_NAME=myapp

build:
	go build -o $(BINARY_NAME)

clean:
	go clean
	rm $(BINARY_NAME)

test: build run

.PHONY: build run clean test
