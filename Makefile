.PHONY: help run build test lint docker-up docker-down clean e2e-up e2e-down e2e-test

test:
	go test -v ./...

lint:
	golangci-lint run ./...

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

e2e-up:
	docker-compose -f docker-compose.e2e.yml up -d

e2e-down:
	docker-compose -f docker-compose.e2e.yml down -v --remove-orphans

e2e-test: e2e-down e2e-up
	@echo "Waiting for services to be ready..."
	@sleep 5
	go test -v ./tests/e2e/...
	@$(MAKE) e2e-down
