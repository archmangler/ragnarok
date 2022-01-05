#!/bin/bash
#Quick AKS deploy script for PoC purposes
#(NOT PRODUCTION WORKLOADS)
#


#terraform
printf "Installing terraform ..."
brew tap hashicorp/tap
brew install tfenv
tfenv install 0.14.11
terraform -v
tfenv use 0.14.11

#mac OSX
brew install azure-cli
az login
#git clone https://github.com/hashicorp/learn-terraform-provision-aks-cluster

cd terraform
printf "Generating AD AppID SP Account ..."

credentials=$(az ad sp create-for-rbac --skip-assignment)
printf "$credentials\n"

#Deploy using terraform ...
terraform init
terraform plan -out terraform.plan
terraform apply terraform.plan

#get the kubeconf
az aks get-credentials --resource-group $(terraform output -raw resource_group_name) --name $(terraform output -raw kubernetes_cluster_name)
