# Quick Redis Deployment for Kubernetes (AKS)

# Install

./deploy.sh

# Configuration and Use

```
"azure-marketplace" has been added to your repositories
NAME: my-release
LAST DEPLOYED: Wed Jan 12 12:32:18 2022
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
CHART NAME: redis
CHART VERSION: 15.7.2
APP VERSION: 6.2.6

** Please be patient while the chart is being deployed **

Redis&trade; can be accessed on the following DNS names from within your cluster:

    my-release-redis-master.default.svc.cluster.local for read/write operations (port 6379)
    my-release-redis-replicas.default.svc.cluster.local for read-only operations (port 6379)



To get your password run:

    export REDIS_PASSWORD=$(kubectl get secret --namespace default my-release-redis -o jsonpath="{.data.redis-password}" | base64 --decode)

To connect to your Redis&trade; server:

1. Run a Redis&trade; pod that you can use as a client:

   kubectl run --namespace default redis-client --restart='Never'  --env REDIS_PASSWORD=$REDIS_PASSWORD  --image marketplace.azurecr.io/bitnami/redis:6.2.6-debian-10-r94 --command -- sleep infinity

   Use the following command to attach to the pod:

   kubectl exec --tty -i redis-client \
   --namespace default -- bash

2. Connect using the Redis&trade; CLI:
   REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli -h my-release-redis-master
   REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli -h my-release-redis-replicas

To connect to your database from outside the cluster execute the following commands:

    kubectl port-forward --namespace default svc/my-release-redis-master 6379:6379 &
    REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli -h 127.0.0.1 -p 6379
```


#Interacting with Redis vi simple cli operations

```
(base) welcome@Traianos-MacBook-Pro redis-storage % 
(base) welcome@Traianos-MacBook-Pro redis-storage % kubectl run --namespace default redis-client --restart='Never'  --env REDIS_PASSWORD=$REDIS_PASSWORD  --image marketplace.azurecr.io/bitnami/redis:6.2.6-debian-10-r94 --command -- sleep infinity
pod/redis-client created
(base) welcome@Traianos-MacBook-Pro redis-storage % 
(base) welcome@Traianos-MacBook-Pro redis-storage % 
(base) welcome@Traianos-MacBook-Pro redis-storage % 
   kubectl exec --tty -i redis-client \
   --namespace default -- bash
I have no name!@redis-client:/$ 
I have no name!@redis-client:/$ 
I have no name!@redis-client:/$ 
I have no name!@redis-client:/$  REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli -h my-release-redis-master
my-release-redis-master:6379> 
my-release-redis-master:6379> 
my-release-redis-master:6379> quit
I have no name!@redis-client:/$ 
I have no name!@redis-client:/$ 
I have no name!@redis-client:/$    REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli -h my-release-redis-replicas
my-release-redis-replicas:6379> 
my-release-redis-replicas:6379> 
my-release-redis-replicas:6379> quit
I have no name!@redis-client:/$ 
```



# References

- https://bitnami.com/stack/redis/helm

