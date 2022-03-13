#!/bin/bash

function install_prerequisites () {
  brew install awscli
  #force it ...
  brew link --overwrite awscli
}

function install_authenticator () {
   brew install aws-iam-authenticator
}

function eks_deploy () {

 printf "initializing EKS terraform state ..."

 terraform init
 terraform plan -out terraform.plan
 terraform apply terraform.plan

}

function get_kube_credentials () {

  CMD="aws eks --region $(terraform output -raw region) update-kubeconfig --name $(terraform output -raw cluster_name)"
  printf "Getting kubeconf with: aws eks --region $(terraform output -raw region) update-kubeconfig --name $(terraform output -raw cluster_name)"
  OUT=$($CMD)
  printf "$OUT\n"

}

function health_checks () {

  printf "=================== Basic Cluster Checks =================\n"
  kubectl get nodes -o wide
  kubectl get ns
  printf "==========================================================\n"

}


function install_metrics_service () {
  printf "metrics server ... coming soon ...\n"
  kubectl apply -f metrics-server-components.yaml
}

#
install_prerequisites
install_authenticator
eks_deploy
health_checks
get_kube_credentials
health_checks
install_metrics_service
