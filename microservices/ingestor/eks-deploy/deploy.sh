#!/bin/bash
#TODO: sed the manifest to update the ingress LB IP for grafana URL
kubectl apply -f ingestor-manifest-eks.yaml
for i in  `ls ../rbac-config/*.yaml`
do
  kubectl apply -f ../rbac-config/$i
done
