#!/bin/bash
#Install kafka and setup all required topics for load-testing service

#kubectl create namespace kafka
function install_requirements () {
  brew install helm
  helm repo add incubator https://charts.helm.sh/incubator
  #deploy the test client for kafka troubleshooting
  kubectl create -f kafka-client.yaml
}

function install_kafka_from_charts() {
  helm install kafka incubator/kafka -f values.yaml --namespace kafka
}

function boot_delay () {
  #delay loop to give kafka cluster a chance to boot up and synchronise
  for i in `seq 1 200`
  do
  echo "`date` waiting for kafka cluster to synch ..."
  kubectl get pods -n kafka
  sleep 2
  done
}

function scale_brokers () {
  #scale the kafka broker count to 14
  kubectl -n kafka scale statefulsets kafka --replicas=14
  #scale the kafka zookeeper count to 5
  #kubectl -n kafka scale statefulsets kafka-zookeeper --replicas=3
}

function create_topics () {
  for topic in messages deadLetter metrics
  do
  kubectl -n kafka exec testclient -- ./bin/kafka-topics.sh --zookeeper kafka-zookeeper:2181 --topic $topic --create --partitions 7 --replication-factor 1
  done
}

function update_topics () {
  #beyond the default, scale messages partitions to 14
  kubectl -n kafka exec testclient -- ./bin/kafka-topics.sh --zookeeper kafka-zookeeper:2181 --topic messages --alter --partitions 14
}

#do it
#install_requirements
install_kafka_from_charts
boot_delay
scale_brokers
create_topics

#update_topics
