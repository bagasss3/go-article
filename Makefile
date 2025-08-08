mock:
	mockgen -source=pkg/model/author.go -destination=internal/mocks/author_mock.go -package=mocks
	mockgen -source=pkg/model/article.go -destination=internal/mocks/article_mock.go -package=mocks

test-unit:
	go test ./internal/service/... ./internal/repository/... ./internal/api/http/... -v -cover -short

run-docker:
	docker-compose up --build

migrate:
	go run main.go migrate

migrate-down:
	go run main.go migrate --direction=down