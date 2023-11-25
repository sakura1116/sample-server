update:
	go get -u -d -v -t ./...
	go mod tidy
deps:
	go mod download
fmt:
	gofmt -w .

run:
	go run main.go

