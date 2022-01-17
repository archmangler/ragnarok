#!/bin/bash
#Quick AKS deploy script for PoC purposes
#(NOT PRODUCTION WORKLOADS)

#terraform
printf "Installing terraform and tfenv on Mac OSX..."
brew tap hashicorp/tap
brew install tfenv
tfenv install 0.14.11
terraform -v
tfenv use 0.14.11

#mac OSX
printf "Install azure cli on Mac OSX ...\n"
brew install azure-cli
az login

cd terraform

#TODO
#printf "Generating AD AppID SP Account ...\n"
#credentials=$(az ad sp create-for-rbac --skip-assignment)
#printf "$credentials\n"

#Deploy using terraform ...
terraform init
terraform plan -out terraform.plan
terraform apply terraform.plan

#get the kubeconf
printf "Updating the kubeconfig ..."
az aks get-credentials --resource-group $(terraform output -raw resource_group_name) --name $(terraform output -raw kubernetes_cluster_name)
