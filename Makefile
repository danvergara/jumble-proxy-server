TEST_URL ?= https%3A%2F%2Fgithub.com%2Fdaywalker90%2Fcln-nip47

.PHONY: docker-build
## docker-build: Builds the Docker image
docker-build:
	@docker build -t jumble-proxy-server .

.PHONY: build
## build: Builds the Go program
build:
	@CGO_ENABLED=0 go build -o bin/jumble-proxy-server .

.PHONY: curl
## curl: Run curl to test the endpoint 
curl:
	curl -X GET http://localhost:8080/sites/$(TEST_URL)

.PHONY: docker-run
## docker-run: Run the container
docker-run: docker-build
	@docker run --rm -e PORT=8080 -p 8080:8080 jumble-proxy-server:latest

.PHONY: test
## test: Runs the tests
test:
	@go test -v -race ./...

.PHONY: help
## help: Prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
