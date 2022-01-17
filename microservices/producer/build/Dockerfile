# syntax=docker/dockerfile:1

FROM golang:1.16-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download
COPY *.go ./

RUN go build -o /producer

RUN mkdir -p /datastore/
RUN mkdir -p /datastore/processed/
RUN mkdir -p /applogs/

RUN ln -sf /dev/stdout /applogs/producer.log

CMD /producer
