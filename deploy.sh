#!/bin/bash
#made for mac osx. deal with it.

function create_namespaces () {
  for ns in ragnarok redis pulsar
  do
    echo "creating namespace $ns"
    kubectl create ns $ns
  done
}

function deploy_aks_cluster {
  printf "deploying kubernetes cluster on Azure Cloud\n"
  mycwd=`pwd`
  cd kubernetes/aks-deploy/
  ./deploy.sh
  cd $mycwd
}

function deploy_ingress_controller(){
  printf "deploy ingress controller ..."
  mycwd=`pwd`
  cd networking
  ./deploy.sh
  cd $mycwd
}

function deploy_prometheus_services () {
  mycwd=`pwd`
  cd monitoring/prometheus/common
  ./deploy.sh
  cd $mycwd
}

function deploy_grafana_services () {
  mycwd=`pwd`
  cd monitoring/grafana
  ./deploy.sh
  cd $mycwd
}

function deploy_pulsar_services () {
  mycwd=`pwd`
  cd pulsar/aks-deploy/apache-pulsar
  ./deploy.sh
  cd $mycwd
}

function deploy_kafka_services () {
  mycwd=`pwd`
  cd kafka-cp
  ./deploy.sh
  cd $mycwd
}

function deploy_redis_services () {
  mycwd=`pwd`
  cd microservices/storage/redis-storage/
  ./deploy.sh
  cd $mycwd
}

function deploy_ingress_service () {
  mycwd=`pwd`
  cd microservices/ingress/aks-nginx-ingress
  ./deploy.sh
  cd $mycwd
}

function deploy_sink_service () {
  mycwd=`pwd`
  cd microservices/load-sink
  ./deploy.sh
  cd $mycwd
}

function deploy_producer_service_aks () {
  mycwd=`pwd`
  cd microservices/producer/aks-deploy/
  ./deploy.sh
  cd $mycwd
}

function deploy_consumer_service_aks () {
  mycwd=`pwd`
  cd microservices/consumer/aks-deploy/
  ./deploy.sh
  cd $mycwd
}

function deploy_loader_service_aks () {
  mycwd=`pwd`
  cd microservices/loader/aks-deploy/
  ./deploy.sh
  cd ../rbac-config
  ./deploy.sh
  cd $mycwd
}

function deploy_local_storage_aks () {
  mycwd=`pwd`
  cd microservices/storage/azurefile-storage/ && ./deploy.sh
  cd $mycwd
}

function update_registry_access () {
  mycwd=`pwd`
  cd microservices/producer/build/ && ./create-secret.sh
  cd $mycwd
}

#Deployment to Azure Cloud
deploy_aks_cluster
create_namespaces
deploy_ingress_controller
deploy_prometheus_services
deploy_grafana_services
deploy_redis_services
#deploy_kafka_services
deploy_pulsar_services
deploy_sink_service
deploy_local_storage_aks
update_registry_access
deploy_producer_service_aks
deploy_consumer_service_aks
deploy_loader_service_aks
deploy_ingress_service
