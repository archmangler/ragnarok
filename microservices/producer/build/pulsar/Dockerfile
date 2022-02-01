# syntax=docker/dockerfile:1

FROM golang:1.16-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

#Pulsar pre-requisites in apk
RUN apk add gcc
RUN apk add musl-dev

COPY oauth2/ ./oauth2/
COPY pulsar/ ./pulsar/

RUN go mod download
COPY *.go ./

RUN go build -o /producer

RUN mkdir -p /datastore/
RUN mkdir -p /datastore/processed/
RUN mkdir -p /applogs/

RUN ln -sf /dev/stdout /applogs/producer.log

CMD /producer
