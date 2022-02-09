#!/bin/bash
#A simple wrapper script for deploying pulsar on kubernetes
namespace="pulsar"

function install_requirements() {

   OUT=$(helm repo add apache https://pulsar.apache.org/charts)
   printf "$OUT\n"
   OUT=$(helm repo update)
   printf "$OUT\n"

}

function deploy_pulsar () {

   helm install pulsar apache/pulsar \
     --timeout 10m \
     --set initialize=true \
     --namespace ${namespace} \
     -f pulsar.yaml

  for i in `seq 1 10`
  do
    kubectl get services -n ${namespace}
    sleep 2
  done

}

function create_default_topics {

   printf "creating and configuring default topics ...\n"
 
   for i in `seq 1 10`
   do
     kubectl get pods -n pulsar| egrep toolset
     sleep 5
   done

   kubectl -n pulsar exec -it  pulsar-toolset-0 -- /pulsar/bin/pulsar-admin tenants create ragnarok
   sleep 5
   kubectl -n pulsar exec -it  pulsar-toolset-0 -- /pulsar/bin/pulsar-admin namespaces create ragnarok/transactions
   sleep 5
   kubectl -n pulsar exec -it  pulsar-toolset-0 -- /pulsar/bin/pulsar-admin namespaces list ragnarok
   sleep 5
   kubectl -n pulsar exec -it  pulsar-toolset-0 -- /pulsar/bin/pulsar-admin topics create persistent://ragnarok/transactions/requests
   sleep 5
   kubectl -n pulsar exec -it  pulsar-toolset-0 -- /pulsar/bin/pulsar-admin topics list ragnarok/transactions

}

install_requirements
deploy_pulsar
create_default_topics
