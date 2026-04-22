.PHONY: run stop test

run:
	docker-compose up --build

stop:
	docker-compose down

test:
	go test -v -cover ./internal/...