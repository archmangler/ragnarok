# syntax=docker/dockerfile:1

FROM golang:1.16-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./

#Pulsar pre-requisites in apk
RUN apk add gcc
RUN apk add musl-dev

RUN go build -o /ingestor

RUN mkdir -p /datastore
RUN mkdir -p /processed
RUN mkdir -p /backups
RUN mkdir -p /applogs
RUN mkdir -p /staging

#install kubectl for kube-ops
RUN apk add curl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.18.0/bin/linux/amd64/kubectl
RUN chmod +x ./kubectl
RUN mv ./kubectl /usr/local/bin/kubectl
RUN ln -sf /dev/stdout /applogs/ingestor.log
RUN apk add --no-cache \
        python3 \
        py3-pip \
    && pip3 install --upgrade pip \
    && pip3 install --no-cache-dir \
        awscli \
    && rm -rf /var/cache/apk/*
RUN aws --version 
CMD /ingestor
