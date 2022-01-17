# syntax=docker/dockerfile:1

FROM golang:1.16-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./

RUN go build -o /loader

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

#install kafkacat / kcat for kafa message extraction
RUN apk add kcat

RUN ln -sf /dev/stdout /applogs/loader.log

CMD /loader
