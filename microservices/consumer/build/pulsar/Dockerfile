# syntax=docker/dockerfile:1

FROM golang:1.16-alpine
WORKDIR /app

RUN apk add gcc
RUN apk add musl-dev

COPY go.mod ./
COPY go.sum ./

COPY oauth2/ ./oauth2/
COPY pulsar/ ./pulsar/
RUN go mod download
COPY *.go ./

RUN go build -o /consumer

RUN mkdir -p /data

CMD /consumer
