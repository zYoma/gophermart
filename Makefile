
run:
	go run cmd/gophermart/main.go -d postgresql://postgres:postgres@127.0.0.1:9432/gophermart

test:
	go test ./...

mock:
	cd internal && mockery --all && cd -

makemigrate:
	goose -dir ./internal/storage/migrations create $(name) sql