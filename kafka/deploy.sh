#!/bin/bash
#Install kafka and setup all required topics for load-testing service

#kubectl create namespace kafka
brew install helm
helm repo add incubator https://charts.helm.sh/incubator
helm install kafka incubator/kafka -f values.yaml --namespace kafka

#deploy the test client for kafka troubleshooting
kubectl create -f kafka-client.yaml

#delay loop to give kafka cluster a chance to boot up and synchronise
for i in `seq 1 100`;
do 
  echo "`date` waiting for kafka cluster to synch ..."
  sleep 2
done

for topic in messages deadLetter metrics
do
  kubectl -n kafka exec testclient -- ./bin/kafka-topics.sh --zookeeper kafka-zookeeper:2181 --topic $topic --create --partitions 7 --replication-factor 1
done

kafka get pods -n kafka
