include .env
export $(shell sed 's/=.*//' .env)

test:
	./scripts/validate-license.sh
	go fmt ./cmd/...
	go test --race ./cmd/...
	go mod tidy
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run -v
run:
	go run --race ./cmd -log.level=DEBUG -topics-file=kafka-topics.yaml $(args)
build:
	docker build --pull . -t paskalmaksim/kafka-topic-admin:dev
push:
	docker push paskalmaksim/kafka-topic-admin:dev
scan:
	trivy image \
	-ignore-unfixed --no-progress --severity HIGH,CRITICAL \
	paskalmaksim/kafka-topic-admin:dev