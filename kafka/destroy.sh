#!/bin/bash
#update/refresh settings from values.yaml

printf "removing kafka helm chart ..."
helm uninstall kafka incubator/kafka --namespace kafka

#wait a while until things are removed
for i in `seq 1 10`
do
  sleep 2
 echo "`date` waiting for kafka app removal to complete ..."
done

#totally purge all pods in kafka
for i in `kubectl get pods -n kafka | awk -s '{print $1}'| egrep -v "NAME"`
do
  kubectl delete pod $i -n kafka
  sleep 10
done

for i in `kubectl get jobs -n kafka | awk -s '{print $1}'| egrep -v "NAME"`
do
  kubectl delete job $i -n kafka
  sleep 10
done

#totally purge all pvc in kafka
for i in `kubectl get pvc -n kafka | awk -s '{print $1}'| egrep -v "NAME"`
do
  printf "deleting pvc $i\n"
  kubectl delete pvc $i -n kafka
done
