#!/bin/bash
#convenient helm insall from Azure Marketplace
#helm repo add azure-marketplace https://marketplace.azurecr.io/helm/v1/repo
#helm install ragnarok azure-marketplace/redis --namespace ragnarok
helm install ragnarok charts/redis --namespace ragnarok
