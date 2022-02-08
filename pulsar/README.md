# Pulsar Messaging Platform

- An alternative to Kafka, said to be more performant and capable.

# Troubleshooting/Validating the installation

```
kubectl -n pulsar exec -it  pulsar-toolset-0 -- /bin/bash

docker exec -it pulsar /pulsar/bin/pulsar-admin clusters list
"standalone"
 
docker exec -it pulsar /pulsar/bin/pulsar-admin tenants list
"public"
"sample"
 
docker exec -it pulsar /pulsar/bin/pulsar-admin tenants create manning
 
docker exec -it pulsar /pulsar/bin/pulsar-admin tenants list
"manning"
"public"
"sample"
 
docker exec -it pulsar /pulsar/bin/pulsar-admin namespaces create manning/chapter03
 
docker exec -it pulsar /pulsar/bin/pulsar-admin namespaces list manning    
"manning/chapter03"
 
docker exec -it pulsar /pulsar/bin/pulsar-admin topics create persistent://manning/chapter03/example-topic
 
docker exec -it pulsar /pulsar/bin/pulsar-admin topics list manning/chapter03
"persistent://manning/chapter03/example-topic"

```


# References

1. Understanding Pulsar			    : https://segmentfault.com/a/1190000041096450/en
2. Why Pulsar is better than Kafka	: https://pandio.com/blog/pulsar-vs-kafka/
3. Helm Installation on Kubernetes	: https://github.com/kafkaesque-io/pulsar-helm-chart
4. Interacting with Pulsar		    : https://livebook.manning.com/book/pulsar-in-action/chapter-3/v-9/1



