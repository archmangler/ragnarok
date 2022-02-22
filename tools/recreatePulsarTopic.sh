#!/bin/bash
#Utility script to recreate a corrupted pulsar topic
#

TOPIC_NAME="persistent://ragnarok/transactions/requests"
PULSAR_NAMESPACE="pulsar"
TOPIC_CREATE_COMMAND="kubectl -n ${PULSAR_NAMESPACE} exec -it pulsar-toolset-0 -- /pulsar/bin/pulsar-admin topics create ${TOPIC_NAME}"
TOPIC_LIST_COMMAND="kubectl -n ${PULSAR_NAMESPACE} exec -it pulsar-toolset-0 -- /pulsar/bin/pulsar-admin topics list ragnarok/transactions"
BROKER_RESTART_CMD="kubectl rollout restart sts pulsar-broker --namespace ${PULSAR_NAMESPACE}"
BROKER_RESTART_CHECK="kubectl get pods -n ${PULSAR_NAMESPACE} | egrep 'pulsar-broker'| awk '{print \$5}' | egrep '^[0-9]+s'"

function restart_brokers () {
   printf "Running ${BROKER_RESTART_CMD}\n"
   OUT=$(${BROKER_RESTART_CMD})
   printf "restarting broker ...\n"
   printf "$OUT\n"
   for i in `seq 1 10`
   do
     printf "Running: ${BROKER_RESTART_CHECK}\n"
     OUT=$(kubectl get pods -n ${PULSAR_NAMESPACE} | egrep 'pulsar-broker'| awk '{print $5}' | egrep '^[0-9]+s')
     echo $OUT
     sleep 5
   done
}

function unload_topic () {

   OUT=$(kubectl -n ${PULSAR_NAMESPACE} exec -it pulsar-toolset-0 -- /pulsar/bin/pulsar-admin topics unload persistent://ragnarok/transactions/requests)
   printf "$OUT\n"

   OUT=$(kubectl -n ${PULSAR_NAMESPACE} exec -it pulsar-toolset-0 -- /pulsar/bin/pulsar-admin topics list ragnarok/transactions)
   printf "$OUT\n"

}

#restart_brokers
unload_topic
