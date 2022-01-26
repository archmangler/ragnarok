#!/bin/bash
#Quick AKS deploy script for PoC purposes
#(NOT PRODUCTION WORKLOADS)

function  install_macosx_requirements () {

    printf "Installing terraform and tfenv on Mac OSX..."
    brew tap hashicorp/tap
    brew install tfenv
    tfenv install 0.14.11
    terraform -v
    tfenv use 0.14.11

    #for json mangling ...
    brew install jq
    
    printf "Install azure cli on Mac OSX ...\n"
    brew install azure-cli
    az login --use-device-code
    
    mkdir -p terraform
    cd terraform

}

function generate_azure_appid () {

    #TODO: Automate creation and revocation with time-limited credentials -> integrate a secrets manager
    #az ad sp create-for-rbac --name test-aks-sp --skip-assignment
    printf "Generating AD AppID SP Account ...\n"
    credentials=$(az ad sp create-for-rbac --name ${DEPLOYMENT_NAME} --skip-assignment)
    printf "$credentials\n"

    appId=$(echo $credentials | jq -r '.appId')
    password=$(echo $credentials | jq -r '.password')

    #write to  file for terraform vars
    printf "appId=\"$appId\"\n" > terraform.tfvars
    printf "password=\"$password\"\n" >> terraform.tfvars

}

function deploy_infrastructure () {

    #Deploy using terraform ...
    terraform init
    terraform plan -out terraform.plan
    terraform apply terraform.plan

}

function get_kubeconfig () {


    #some delay (TODO: do active polling instead!)
    for i in `seq 1 100`
    do
      printf "waiting for kubernetes cluster to boot up ...\n"
      nsData=$(kubectl get ns)
      printf "$nsData\n"
      sleep 2
    done

    #get the kubeconf
    printf "Updating the kubeconfig ..."
    az aks get-credentials --resource-group $(terraform output -raw resource_group_name) --name $(terraform output -raw kubernetes_cluster_name)


}

#install basic tooling + terraform version
install_macosx_requirements 

#get the azure application credentials
generate_azure_appid

#deploy El Cluster!
deploy_infrastructure

#check for signs of life then update kubeconf...
get_kubeconfig
