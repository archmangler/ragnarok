#!/bin/bash
#A simple wrapper script for deploying pulsar on kubernetes

function install_requirements() {
   OUT=$(helm repo add kafkaesque https://helm.kafkaesque.io)
   printf "$OUT\n"
   OUT=$(helm repo update)
   printf "$OUT\n"
}

function deploy_pulsar () {
  OUT=$(helm install pulsar kafkaesque/pulsar --namespace pulsar --create-namespace --values storage_values.yaml)
  printf "$OUT\n"
}

install_requirements
deploy_pulsar
