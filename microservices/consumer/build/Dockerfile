# syntax=docker/dockerfile:1

FROM golang:1.16-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download
COPY *.go ./
COPY use.py ./
COPY sign384.py ./

RUN go build -o /consumer
RUN mkdir -p /data

CMD /consumer
