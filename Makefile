.PHONY: docker-build
## docker-build: Builds the Docker image
docker-build:
	@docker build -t jumble-proxy-server .

.PHONY: build
## build: Builds the Go program
build:
	@CGO_ENABLED=0 go build -o bin/jumble-proxy-server .

.PHONY: docker-run
## docker-run: Run the container
docker-run: docker-build
	@docker run --rm -e ALLOW_ORIGIN=https://jumble.social -e PORT=8080 -p 8080:8080 jumble-proxy-server:latest

.PHONY: test
## test: Runs the tests
test:
	go test -v -race ./...

.PHONY: help
## help: Prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
