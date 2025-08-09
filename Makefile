.PHONY: run run_full

run:
	go run cmd/api/main.go

run_full:
	docker-compose up -d


