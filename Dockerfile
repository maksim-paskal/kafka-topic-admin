FROM golang:1.17-alpine as build

COPY ./cmd /usr/src/kafka-topic-creation/cmd
COPY go.* /usr/src/kafka-topic-creation/
COPY .git /usr/src/kafka-topic-creation/

ENV GOOS=linux
ENV GOARCH=amd64
ENV GOFLAGS="-trimpath"

RUN apk add --no-cache gcc musl-dev

RUN apk add --no-cache gcc musl-dev \
  && cd /usr/src/kafka-topic-creation \
  && go mod download \
  && go mod verify \
  && go build -ldflags='-w -s -extldflags "-static"' -tags musl -v -o kafka-topic-creation ./cmd

RUN /usr/src/kafka-topic-creation/kafka-topic-creation -h

FROM alpine:latest

COPY --from=build /usr/src/kafka-topic-creation/kafka-topic-creation /app/kafka-topic-creation

WORKDIR /app

RUN apk upgrade \
&& addgroup -g 101 -S app \
&& adduser -u 101 -D -S -G app app

USER 101

ENTRYPOINT [ "/app/kafka-topic-creation" ]