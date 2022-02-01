#!/bin/bash
#update/refresh settings from values.yaml
namespace="pulsar"

printf "removing pulsar helm chart ..."
helm uninstall pulsar apache/pulsar --namespace ${namespace}

#wait a while until things are removed
for i in `seq 1 30`
do
  sleep 2
  echo "`date` waiting for pulsar app removal to complete ..."
  kubectl get pvc -n ${namespace}
done

#totally purge all pods in pulsar
for i in `kubectl get pods -n ${namespace} | egrep "pulsar" | awk -s '{print $1}'| egrep -v "NAME"`
do
  kubectl delete pod $i -n ${namespace} 
  sleep 10
done

for i in `kubectl get jobs -n ${namespace} | egrep "pulsar" | awk -s '{print $1}'| egrep -v "NAME"`
do
  kubectl delete job $i -n ${namespace} 
  sleep 10
done

#totally purge all pvc in pulsar
for i in `kubectl get pvc -n ${namespace} |egrep "pulsar" |  awk -s '{print $1}'| egrep -v "NAME"`
do
  printf "deleting pvc $i\n"
  kubectl delete pvc $i -n ${namespace}
done

for i in `kubectl get secrets -n ${namespace} |egrep "pulsar" | awk -s '{print $1}'| egrep -v "NAME"`
do
  printf "deleting secret $i\n"
  kubectl delete secret $i -n ${namespace}
done
