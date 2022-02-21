#!/bin/bash
#simple deploy to EKS
#

kubectl create namespace ragnarok
kubectl apply -f producer-manifest-eks-sts.yaml
