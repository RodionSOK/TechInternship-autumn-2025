.PHONY: help run build test lint docker-up docker-down clean

test:
	go test -v ./...

lint:
	golangci-lint run ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down
