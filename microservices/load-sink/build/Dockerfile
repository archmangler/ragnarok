# syntax=docker/dockerfile:1

FROM golang:1.16-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download
COPY *.go ./

RUN go build -o /anvil
RUN mkdir -p /processed

RUN ln -sf /dev/stdout /var/log/load-sink.log

CMD /anvil
