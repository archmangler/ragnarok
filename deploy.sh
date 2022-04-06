#!/bin/bash
#made for mac osx. deal with it.

function create_namespaces () {
  for ns in ragnarok redis pulsar
  do
    echo "creating namespace $ns"
    kubectl create ns $ns
  done
}

function deploy_kubernetes_cluster(){
  printf "deploying  kubernetes ..."
  if [[ $target_cloud == *"azure"* ]]
  then
    printf "Deploying an AKS cluster ...\n"
    deploy_aks_cluster
  fi
  if [[ $target_cloud == *"aws"* ]]
  then
    printf "Deploying an EKS cluster ...\n"
    deploy_eks_cluster
  fi
}

function deploy_eks_cluster {
  printf "deploying kubernetes cluster on AWS Cloud\n"
  mycwd=`pwd`
  cd kubernetes/eks-deploy/
  ./deploy.sh
  cd $mycwd
}

function deploy_aks_cluster {
  printf "deploying kubernetes cluster on Azure Cloud\n"
  mycwd=`pwd`
  cd kubernetes/aks-deploy/
  ./deploy.sh
  cd $mycwd
}

function deploy_ingress_controller(){
  printf "deploy ingress controllers ..."
  mycwd=`pwd`

  if [[ $target_cloud == *"azure"* ]]
  then
    deploy_aks_ingress_controller
  fi
  if [[ $target_cloud == *"aws"* ]]
  then
    deploy_eks_ingress_controller
  fi 
  cd $mycwd
}

function deploy_aks_ingress_controller(){
  printf "deploy AKS ingress controller ..."
  mycwd=`pwd`
  cd networking/aks-deploy
  ./deploy.sh
  cd $mycwd
}

function deploy_eks_ingress_controller(){
  printf "deploy EKS ingress controller ..."
  mycwd=`pwd`
  cd networking/eks-deploy
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
  if [[ $target_cloud == *"azure"* ]]
  then
    cd microservices/storage/redis-storage/aks-deploy
    ./deploy.sh
  fi
  if [[ $target_cloud == *"aws"* ]]
  then
    cd microservices/storage/redis-storage/eks-deploy
    ./deploy.sh
  fi
  cd $mycwd
}

function deploy_ingress_service () {
  mycwd=`pwd`
  if [[ $target_cloud == *"azure"* ]]
  then
    cd microservices/ingress/aks-deploy
    ./deploy.sh
  fi
  if [[ $target_cloud == *"aws"* ]]
  then
    cd microservices/ingress/eks-deploy
    ./deploy.sh
  fi
  cd $mycwd
}

function deploy_sink_service () {
  mycwd=`pwd`
  if [[ $target_cloud == *"azure"* ]]
  then
    cd microservices/load-sink/aks-deploy
    ./deploy.sh
  fi

  if [[ $target_cloud == *"aws"* ]]
  then
    cd microservices/load-sink/eks-deploy
    ./deploy.sh
  fi
  cd $mycwd
}

function deploy_producer_service () {
  mycwd=`pwd`
  if [[ $target_cloud == *"azure"* ]]
  then
  cd microservices/producer/aks-deploy/
  ./deploy.sh
  fi
  if [[ $target_cloud == *"aws"* ]]
  then
  cd microservices/producer/eks-deploy/
  ./deploy.sh
  fi

  cd $mycwd
}

function deploy_consumer_service () {
  mycwd=`pwd`
  if [[ $target_cloud == *"azure"* ]]
  then
  cd microservices/consumer/aks-deploy/
  ./deploy.sh
  fi
  if [[ $target_cloud == *"aws"* ]]
  then
  cd microservices/consumer/eks-deploy/
  ./deploy.sh
  fi
  cd $mycwd
}

function deploy_loader_service () {
  mycwd=`pwd`
  if [[ $target_cloud == *"azure"* ]]
  then
    cd microservices/loader/aks-deploy/
    ./deploy.sh
    cd rbac-config
    ./deploy.sh
  fi
  if [[ $target_cloud == *"aws"* ]]
  then
    cd microservices/loader/eks-deploy/
    ./deploy.sh
    cd rbac-config
    ./deploy.sh
  fi
  cd $mycwd
}

function deploy_local_storage () {
  mycwd=`pwd`
  if [[ $target_cloud == *"azure"* ]]
  then
     cd microservices/storage/afs-storage/ && ./deploy.sh
  fi
  if [[ $target_cloud == *"aws"* ]]
  then
     printf "no storage for EKS ...\n"
  fi
  cd $mycwd
}

function update_registry_access () {
  mycwd=`pwd`
  cd microservices/producer/build/ && ./create-secret.sh
  cd $mycwd
}

function deploy_source_data_storage () {
  mycwd=`pwd`
  cd storage/eks-deploy && ./deploy.sh && ./eksS3access.sh
  cd $mycwd
}

#microservices/ingestor/eks-deploy
function deploy_data_ingestor () {
  mycwd=`pwd`
  cd microservices/ingestor/eks-deploy && ./deploy.sh
  cd $mycwd
}

#microservices/ingestor/eks-deploy


#Deployment to Azure Cloud
deploy_kubernetes_cluster
create_namespaces
deploy_ingress_controller
deploy_ingress_service
deploy_prometheus_services
deploy_grafana_services
deploy_redis_services
#make this selectable at some point!
#deploy_kafka_services
deploy_pulsar_services
deploy_local_storage
update_registry_access
deploy_sink_service
deploy_producer_service
deploy_consumer_service
deploy_loader_service
deploy_source_data_storage
deploy_data_ingestor
