.PHONY: build run test clean

build:
	go build -o gook

run: build
	./gook

test:
	go test -v ./...

clean:
	rm -f gook
	go clean

lint:
	golangci-lint run