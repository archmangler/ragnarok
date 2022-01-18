#!/bin/bash
export AKS_RESOURCE_GROUP="anvil-mjolner-rg"
export AKS_CLUSTER_NAME="anvil-mjolner-benchmarking-cluster"
export PUBLIC_IP_SKU="Standard"
export IP_ALLOCATION_METHOD="static"
export DEPLOYMENT_NAME="${AKS_CLUSTER_NAME}_SP"
