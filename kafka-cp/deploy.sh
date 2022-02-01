#!/bin/bash
#Install kafka and setup all required topics for load-testing service

#kubectl create namespace kafka
function install_requirements () {
  brew install helm
  #deploy the test client for kafka troubleshooting
  kubectl create -f kafka-client.yaml
}

function install_kafka_from_charts() {
  helm install kafka cp-helm-charts -f values.yaml --namespace kafka
}

function boot_delay () {
  #make this "intelligent"
  #delay loop to give kafka cluster a chance to boot up and synchronise
  for i in `seq 1 100`
  do
    echo "`date` waiting for kafka cluster to synch ..."
    kubectl get pods -n kafka
    check_kafka_topic_health
    sleep 2
  done
}

function create_topics () {
  for topic in messages
  do
    kubectl -n kafka exec kafka-client -- kafka-topics --zookeeper kafka-cp-zookeeper-headless:2181 --topic $i --create --partitions 3 --replication-factor 1 --if-not-exists
  done
}

function check_kafka_topic_health () {
    printf "\n"
    OUT=$(kubectl -n kafka exec kafka-client -- kafka-topics --zookeeper kafka-cp-zookeeper-headless:2181 --describe)
    printf "$OUT\n"
}

function update_topics () {
  #beyond the default, scale messages partitions to whatever ...
  OUT=$(kubectl -n kafka exec kafka-client -- kafka-topics --zookeeper kafka-cp-zookeeper-headless:2181 --topic messages --alter --partitions 28)
  check_kafka_topic_health
}

#deploy all the things ...
install_requirements
install_kafka_from_charts
boot_delay
create_topics
update_topics
boot_delay
