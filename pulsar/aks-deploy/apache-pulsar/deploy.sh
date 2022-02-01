#!/bin/bash
#A simple wrapper script for deploying pulsar on kubernetes
namespace="pulsar"

function install_requirements() {
   OUT=$(helm repo add apache https://pulsar.apache.org/charts)
   printf "$OUT\n"
   OUT=$(helm repo update)
   printf "$OUT\n"
}

function deploy_pulsar () {
  
   helm install pulsar apache/pulsar \
     --timeout 10m \
     --set initialize=true \
     --namespace ${namespace} \
     -f pulsar.yaml

  for i in `seq 1 10`
  do
    kubectl get services -n ${namespace}
    sleep 2
  done
}

install_requirements
deploy_pulsar
