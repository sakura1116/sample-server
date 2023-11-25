DOCKER_COMPOSE := docker-compose -f docker-compose.yml

docker/setup:
	$(DOCKER_COMPOSE) up -d

update:
	go get -u -d -v -t ./...
	go mod tidy
deps:
	go mod download
fmt:
	gofmt -w .

run:
	go run main.go

