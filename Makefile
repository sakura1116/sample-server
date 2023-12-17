DOCKER_COMPOSE := docker-compose -f docker-compose.yml
export BUILD_TARGET=development

docker/setup:
	$(DOCKER_COMPOSE) up -d

docker/db/ssh:
	$(DOCKER_COMPOSE) exec db /bin/bash

docker/db/cli:
	$(DOCKER_COMPOSE) exec db mysql -u root -ppassword sample

docker/db/migrate:
	$(DOCKER_COMPOSE) exec batch /usr/bin/make db/migrate -C migrations

docker/api/ssh:
	$(DOCKER_COMPOSE) exec api /bin/bash

docker/batch/ssh:
	$(DOCKER_COMPOSE) exec batch /bin/bash

update:
	go get -u -d -v -t ./...
	go mod tidy
deps:
	go mod download

fmt:
	gofmt -w .

test: fmt
	go test -v -race ./...

bash:
	/bin/bash
