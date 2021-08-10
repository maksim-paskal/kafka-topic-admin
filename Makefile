include .env
export $(shell sed 's/=.*//' .env)

test:
	./scripts/validate-license.sh
	go fmt ./cmd/...
	go test --race ./cmd/...
	go mod tidy
	golangci-lint run -v
run:
	go run --race ./cmd -log.level=DEBUG -topics-file=kafka-topics.yaml $(args)
build:
	docker build . -t paskalmaksim/kafka-topic-admin:dev
push:
	docker push paskalmaksim/kafka-topic-admin:dev