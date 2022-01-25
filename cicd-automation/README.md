# Cloud Deployment Pipeline Automation Components for Load Testing Solution

*TODO*

```
[p] master deloyment script wrapper 
 [p][] `./deploy.sh <destroy|deploy> <eks|aks|local>`
 [x][x][] aks cluster deployer script
 [x][x][] ingress deployment script
 [x][x][] prometheus metrics collector deployment script
 [x][x][] Grafana deployment script
 [x][x][] kafka cluster deployment script
 [x][x][] microservices deployment scripts
  [x][x][] loader/manager service
  [x][x][] load-sink service
  [x][x][] producer service
  [x][x][] consumer service
 [p][] destroy function: `./deploy.sh <destroy|deploy> <eks|aks|local>`
```

---

```
NAME: ragnarok
LAST DEPLOYED: Mon Jan 17 23:22:39 2022
NAMESPACE: ragnarok
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
CHART NAME: redis
CHART VERSION: 15.7.5
APP VERSION: 6.2.6

** Please be patient while the chart is being deployed **

Redis&trade; can be accessed on the following DNS names from within your cluster:

    ragnarok-redis-master.ragnarok.svc.cluster.local for read/write operations (port 6379)
    ragnarok-redis-replicas.ragnarok.svc.cluster.local for read-only operations (port 6379)
    ragnarok-redis-replicas.default.svc.cluster.local:

To get your password run:

    export REDIS_PASSWORD=$(kubectl get secret --namespace ragnarok ragnarok-redis -o jsonpath="{.data.redis-password}" | base64 --decode)

To connect to your Redis&trade; server:

1. Run a Redis&trade; pod that you can use as a client:

   kubectl run --namespace ragnarok redis-client --restart='Never'  --env REDIS_PASSWORD=$REDIS_PASSWORD  --image marketplace.azurecr.io/bitnami/redis:6.2.6-debian-10-r97 --command -- sleep infinity

   Use the following command to attach to the pod:

   kubectl exec --tty -i redis-client \
   --namespace ragnarok -- bash

2. Connect using the Redis&trade; CLI:
   REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli -h ragnarok-redis-master
   REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli -h ragnarok-redis-replicas

To connect to your database from outside the cluster execute the following commands:

    kubectl port-forward --namespace ragnarok svc/ragnarok-redis-master 6379:6379 &
    REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli -h 127.0.0.1 -p 6379
(base) welcome@Traianos-MacBook-Pro ragnarok % 



```


