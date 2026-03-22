.PHONY: run build test lint clean docker-up docker-down

run:
	go run ./cmd/glide

build:
	go build -o bin/glide ./cmd/glide

test:
	go test ./...

test-verbose:
	go test -v ./...

lint:
	go vet ./...

clean:
	rm -rf bin/

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down