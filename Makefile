DOCKER_COMPOSE := docker-compose -f docker-compose.yml
export BUILD_TARGET=development

docker/setup:
	$(DOCKER_COMPOSE) up

docker/db/ssh:
	$(DOCKER_COMPOSE) exec db /bin/bash

docker/db/cli:
	$(DOCKER_COMPOSE) exec db mysql -u root -ppassword sample

docker/api/ssh:
	$(DOCKER_COMPOSE) exec db /bin/bash

update:
	go get -u -d -v -t ./...
	go mod tidy
deps:
	go mod download

fmt:
	gofmt -w .

test: fmt
	go test -v -race ./...
run:
	go run main.go
