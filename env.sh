#!/bin/bash
#Kubernetes cluster settings for AKS or EKS
export target_cloud="aws"
export AWS_DEPLOY_REGION="ap-southeast-1"
export AWS_CLUSTER_NAME="ragnarok-eks-mjollner-poc"
export AWS_ACCOUNT_NUMBER="524513049339"
export AKS_RESOURCE_GROUP="anvil-mjollner-rg"
export AKS_CLUSTER_NAME="anvil-mjollner-benchmarking-cluster"
export PUBLIC_IP_SKU="Standard"
export IP_ALLOCATION_METHOD="static"
export DEPLOYMENT_NAME="${AKS_CLUSTER_NAME}_SP"
