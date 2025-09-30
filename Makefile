GO_TOOLCHAIN ?= go1.22.7
DOCKER_IMAGE ?= gin-balancer:latest
PORT ?= 8080

.PHONY: run test docker-build docker-run

run:
	GOTOOLCHAIN=$(GO_TOOLCHAIN) go run ./main.go

test:
	GOTOOLCHAIN=$(GO_TOOLCHAIN) go test ./...

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build
	docker run --rm -p $(PORT):8080 $(DOCKER_IMAGE)
