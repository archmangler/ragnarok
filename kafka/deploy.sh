#!/bin/bash
#Install kafka and setup all required topics for load-testing service

#kubectl create namespace kafka
brew install helm
helm repo add incubator https://charts.helm.sh/incubator
helm install kafka incubator/kafka -f values.yaml --namespace kafka

#deploy the test client for kafka troubleshooting
kubectl create -f kafka-client.yaml

#delay loop to give kafka cluster a chance to boot up and synchronise
for i in `seq 1 200`;
do 
  echo "`date` waiting for kafka cluster to synch ..."
  kubectl get pods -n kafka
  sleep 2
done

#scale the kafka broker count to 14
kubectl -n kafka scale statefulsets kafka --replicas=14
#scale the kafka zookeeper count to 5 
kubectl -n kafka scale statefulsets kafka-zookeeper --replicas=5

for topic in messages deadLetter metrics
do
  kubectl -n kafka exec testclient -- ./bin/kafka-topics.sh --zookeeper kafka-zookeeper:2181 --topic $topic --create --partitions 7 --replication-factor 1
done

