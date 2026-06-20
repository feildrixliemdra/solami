.PHONY: build clean start test

build:
	go build -o bin/solami main.go

start:
	go run main.go

solami: 
	./bin/solami

test:
	go test ./...

clean:
	rm -rf bin/