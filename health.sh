#!/bin/bash
#simple script to troubleshoot/test the health of the deployment

function check_kubernetes_api_endpoint() {
    RES=$(kubectl get ns)
    echo "$RES"
}

function check_kubernetes_ingress() {
  if [[ $target_cloud == *"azure"* ]]
  then
    EIP=$(kubectl get service ingress-nginx-controller -n ingress-basic -o json| jq -r '.spec.loadBalancerIP')
    echo $EIP
    for i in sink-admin loader-admin sink-orders
    do
      output=$(curl -s "$EIP/$i")
      printf "got service endpoint: $output\n"
    done
  fi
  if [[ $target_cloud == *"aws"* ]]
  then
    for i in `seq 1 5`
    do
      OUT=$(kubectl get ingress -n ragnarok -o json| jq -r '.items|.[]|.status|.loadBalancer.ingress|.[]|.hostname')
      printf "Got ingress external address: $OUT\n"
      sleep 2
      OUT=$(curl -s $OUT/loader-admin)
      printf "connecting to external management endpoint: $OUT\n"
    done
  fi 
}

function text_divider () {
    label=$1
    printf "\n"
    printf "=========================== $label ==========================\n"
    printf "\n"
}

function check_kafka_cluster() {
    OUT=$(kubectl get pods -n kafka)
    printf "$OUT"
}

function check_pulsar_cluster() {
    OUT=$(kubectl get pods -n pulsar)
    printf "$OUT"
}

function check_redis_cluster() {
    OUT=$(kubectl get pods -n redis| egrep -i redis)
    printf "$OUT"
}

function check_producer_pool_health() {

    OUT=$(kubectl get pods -n ragnarok | egrep -i producer | egrep -i "Running"| wc -l)
    printf "Running producers: $OUT\n"
    
    OUT=$(kubectl get pods -n ragnarok | egrep -i producer | egrep -i "Error"| wc -l)
    printf "Errored producers: $OUT\n"
    
    OUT=$(kubectl get pods -n ragnarok | egrep -i producer | egrep -i "Pending"| wc -l)
    printf "Pending producers: $OUT\n"
    
    OUT=$(kubectl get pods -n ragnarok | egrep -i producer | egrep -i "Evicted"| wc -l)
    printf "Evicted producers: $OUT\n"
    
    OUT=$(kubectl get pods -n ragnarok | egrep -i producer | egrep -i "CrashLoopBackOff"| wc -l)
    printf "CrashLoopBackOff producers: $OUT\n"
}

function check_consumer_pool_health() {

    OUT=$(kubectl get pods -n ragnarok | egrep -i consumer | egrep -i "Running"| wc -l)
    printf "Running consumers: $OUT\n"

    OUT=$(kubectl get pods -n ragnarok | egrep -i consumer| egrep -i "Error"| wc -l)
    printf "Errored consumers: $OUT\n"
    
    OUT=$(kubectl get pods -n ragnarok | egrep -i consumer | egrep -i "Pending"| wc -l)
    printf "Pending consumers: $OUT\n"
    
    OUT=$(kubectl get pods -n ragnarok | egrep -i consumer | egrep -i "Evicted"| wc -l)
    printf "Evicted consumers: $OUT\n"

    OUT=$(kubectl get pods -n ragnarok | egrep -i consumer | egrep -i "CrashLoopBackOff"| wc -l)
    printf "CrashLoopBackOff consumers: $OUT\n"
    
}

function check_kafka_topic_health () {
    printf "\n"
    OUT=$(kubectl -n kafka exec kafka-client -- kafka-topics --zookeeper kafka-cp-zookeeper-headless:2181 --describe)
    printf "$OUT\n"
}

#1.

text_divider "checking api endpoint reachable"
check_kubernetes_api_endpoint

#2.
text_divider "getting external load balancer IP"
check_kubernetes_ingress

#3.
text_divider "checking kafka cluster health"
check_kafka_cluster

#4.
text_divider "checking pulsar cluster health"
check_pulsar_cluster

#4.
text_divider "checking redis cluster health"
check_redis_cluster

#5.
text_divider "checking producer pool health"
check_producer_pool_health

#6.
text_divider "checking consumer pool health"
check_consumer_pool_health

#7.
text_divider "checking kafka topic status"
check_kafka_topic_health 

exit 1
