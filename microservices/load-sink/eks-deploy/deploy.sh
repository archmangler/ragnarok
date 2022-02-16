#!/bin/bash
#simple deploy to EKS
#
kubectl delete -f load-sink-eks-deployment.yaml
kubectl apply -f load-sink-eks-deployment.yaml
