.PHONY: run stop test integration-test test-all

run:
	docker-compose up --build

stop:
	docker-compose down

test:
	go test -v -count=1 -cover ./internal/...

integration-test:
	go test -v -count=1 -tags=integration ./internal/storage/postgres/...

