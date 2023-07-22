dev:
	go run cmd/main.go

start:
	docker-compose up --build -d

stop:
	docker compose down