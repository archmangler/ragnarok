#!/bin/bash
#Deploy the ingress controller
#Based on the good work done here: https://stacksimplify.com/azure-aks/azure-kubernetes-service-ingress-basics/

AKS_RESOURCE_GROUP=${AKS_RESOURCE_GROUP}
AKS_CLUSTER_NAME=${AKS_CLUSTER_NAME}
PUBLIC_IP_SKU="Standard"
IP_ALLOCATION_METHOD="static"
INGRESS_NAMESPACE="ingress-basic"

function get_ingress_public_address () {
  # Get the resource group name of the AKS cluster 
  AKS_CLUSTER_RG=$(az aks show \
    --resource-group ${AKS_RESOURCE_GROUP} \
    --name ${AKS_CLUSTER_NAME} \
    --query nodeResourceGroup \
    -o tsv)

  # TEMPLATE - Create a public IP address with the static allocation
  INGRESS_PUBLIC_IP=$(az network public-ip create \
    --resource-group ${AKS_CLUSTER_RG} \
    --name "${AKS_CLUSTER_NAME}_INGRESS_IP" \
    --sku ${PUBLIC_IP_SKU}\
    --allocation-method ${IP_ALLOCATION_METHOD} \
    --query publicIp.ipAddress \
    -o tsv) 
  
    printf "ingress ${IP_ALLOCATION_METHOD} public ip ${INGRESS_PUBLIC_IP}\n"
}

function deploy_ingress_controller () {

  #Assuming we're on on Mac OSX Install Helm3 (if not installed)
  printf "Installing helm on mac osx ...\n"
  brew install helm

  # Create a namespace for your ingress resources
  printf "creating ingress namespace ...\n"
  kubectl create namespace ${INGRESS_NAMESPACE}

  # Add the official stable repository
  helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
  helm repo add stable https://kubernetes-charts.storage.googleapis.com/
  helm repo update

#  Customizing the Chart Before Installing. 
helm show values ingress-nginx/ingress-nginx

# Use Helm to deploy an NGINX ingress controller
helm install ingress-nginx ingress-nginx/ingress-nginx \
    --namespace ingress-basic \
    --set controller.replicaCount=2 \
    --set controller.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set defaultBackend.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set controller.service.externalTrafficPolicy=Local \
    --set controller.service.loadBalancerIP="${INGRESS_PUBLIC_IP}" 

# Replace Static IP captured in Step-02
helm install ingress-nginx ingress-nginx/ingress-nginx \
    --namespace ingress-basic \
    --set controller.replicaCount=2 \
    --set controller.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set defaultBackend.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set controller.service.externalTrafficPolicy=Local \
    --set controller.service.loadBalancerIP="${INGRESS_PUBLIC_IP}" 

# List Services with labels
printf "listing services with lables ..."
kubectl get service -l app.kubernetes.io/name=ingress-nginx --namespace ingress-basic

# List Pods
printf "Listing ingress pods ...\n"
kubectl get pods -n ingress-basic
kubectl get all -n ingress-basic

# Access Public IP
printf "testing public ingress access ..."
curl -v http://${INGRESS_PUBLIC_IP}

# Output should be
# 404 Not Found from Nginx

printf "Verify Load Balancer on Azure Mgmt Console - Primarily refer Settings -> Frontend IP Configuration\n"

}


#Ingress deployment step 1
get_ingress_public_address

#Ingress deployment step #2
deploy_ingress_controller

