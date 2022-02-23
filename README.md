# Table of Contents

* [Description](#LoadGenerationSolutionforTradeMatchingSystems)
* [Operation](#Operation)
* [How it Works](#HowitWorks)
* [Installation](#Installation)
* [Compute Resources](#ComputeResources) 

## Load Generation Solution for Trade Matching Systems

A containerised event-driven system for generating transaction load on trading systems.

![alt text](images/load-testing-scenario-one-implementation.drawio.png?raw=true "System Architecture")

## Operation 

Users select one of the following options from the Test Management Portal:

- Generate a new sequence of json files as load requests in a range 1 to n.
- Load pre-existing messages from a kafka topic (sequence n to m) as inputs to a new load test
- Restart producer / consumer kubernetes pods to trigger the load test run
- Back up files from previous load test (to use in future load test if needed)
- Load historical load-test data for input to new load test
- Access Grafana metrics of previous load tests

![alt text](images/management-portal.png?raw=true "Load Test Management Portal")

Load test activity:

- Upon restart, producer pods consume thy generated data files from "/datastore/" and submit to the "messages" kafka topic.
- Consumer pods read JSON messages from the "messages" topic and submit them as HTTP JSON requests via REST to the target ("load sink")

Producers, consumers and management nodes are implemented as containerised golang services in kubernetes pods.
Each service is instrumented with the following custom metrics for evaluating the load testing activity:

- Messages generated for input
- Messages read and produced to the topic by producers
- Producer errors generated in submitting requests to "messages" topic
- Consumer errors generated in submitting requests to target api
- Rate of message submission to the target API (by consumers)

View the producer, consumer and other system component metrics (including the dummy target application provided) on the metrics dashboard:

![alt text](images/metrics-dashboard.png?raw=true "Metrics Dashboard")

## How it Works

*Refer to the high level diagram:*

![alt text](images/load-engine.drawio.png?raw=true "high level system diagram")

* Load data is generated (by the user) using the load content manager test UI.
* The data is stored in Redis
* The "producers" layer consumes the data from Redis and processes it into the asynchronous queue (kafka or pulsar) 
* The "consumers" layer consumes data from the asynchronous queue, does basic content error checking and drives the message load to the target application via HTTP REST (JSON) or FIX (TODO!)
* The target application accepts the data via REST or FIX and writes it to a local disk
* Each of the core layers are elastically scalable, thereby providing the compute power required to drive messages at high rate and volume to the target application.

## Technology and Implementation

- The current immplementation involves the following technologies:

* Redis:
* Pulsar or Kafka:
* Kubernetes (EKS on AWS|AKS on Azure):
* Golang:

## Installation

How to install via cli:

* Step #1: configure the terraform parameters for eks or aks `kubernetes/eks-deploy/eks-cluster.tf`
* Step #2: configure env.sh, then `source env.sh`
* Step #3: run the master deploy script: `./deploy.sh`
* Step #4: run the health check script: `./health.sh`
* Step #5: review the output of health.sh and troubleshoot any issues.

## Compute_Resources

* As a reference for the kubernetes cluster sizing, the following AKS deployment on Azure provides a rough idea of the minimum resources required for cloud deployment

![alt text](images/basic-resource-requirements-26012022.png?raw=true "Cluster Sizing on Cloud")

* This should result in a cluster able to generate ~ 8000 json http requests per second using the architecture outlined here to a target application.

![alt text](images/basic-performance-benchmark-26012022.png?raw=true "Cluster Performance on Cloud")

## Troubleshooting Procedure

- View the requests received by the dummy target application (The "black box")

*make a manual request:*

```
root@testclient:/opt/kafka# curl -v   -X POST  http://10.42.0.199:80/orders   -H 'Content-Type: application/json'   -d '{"Name":"newOrder", "ID":"78912","Time":"223232113111","Data":"new order", "Eventname":"newOrder"}'
Note: Unnecessary use of -X or --request, POST is already inferred.
*   Trying 10.42.0.199:80...
* Connected to 10.42.0.199 (10.42.0.199) port 80 (#0)
> POST /orders HTTP/1.1
> Host: 10.42.0.199
> User-Agent: curl/7.74.0
> Accept: */*
> Content-Type: application/json
> Content-Length: 98
> 
* upload completely sent off: 98 out of 98 bytes
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Sun, 26 Dec 2021 11:03:22 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.42.0.199 left intact
```

*check the error and request count by viewing the admin portal on the dummy app*

```
root@testclient:/opt/kafka# curl --user "admin:Crypt0N0m1c0n##" http://10.42.0.199:80/admin
<html><h1>Anvil Management Portal</h1></html><html> Total Requests: 2</html><html> Total Errors: 0</html>root@testclient:/opt/kafka# 
```

(you may need to do this from inside a test container if connectivity outside the kubernetes cluster is not configured)


# Development Tasks

```
[p] Scalable producer Pool
 [x] Create a Golang Service
  [?] Create a service for the duration of the run
  [x] Consume formatted input data in concurrent mode.
  [x] Publish to kafka Topic
  [x] move processed files to an output queue
  [] Ensure Effective Scaling
   [] Resolve concurrent access problems
    
   [] Scaling of worker processes
   [] job allocation scaling
    [] job allocation: distribute range of inputs among workers
     [] 1. scan the input folder
     [] 2. divide the batch into ranges among the workers
     [] 3. assign  the task  to the workers to loop over each batch

[p] Implement Scalable consumer pool
 [x] Create main worker service
 [x] Consume from messages queue from Kafka
 [x] Submit message via an API (PoC: write to output files in directory)
 
 [] Ensure Effective Scaling
  [] understand Golang Coroutines and channels properly
   [] seek a better code implementation
    [] The case for worker pools with goroutines: https://brandur.org/go-worker-pool

[] Scale Golang operations Concurrently 
 [] Test Scalability
  [] Configure loads
  [] Extract Metrics
  [] Analyse and Present Metrics

[] Baseline Configuration of System
[] Realtime Control of System
```

# Kafka Configuration

```
messages
metrics
deadLetter
```

# Image Build

- https://docs.docker.com/language/golang/build-images/

Registry login (Assuming we're using ECR):

```
aws ecr get-login-password --region ap-southeast-1 | docker login --username AWS --password-stdin 605125156525.dkr.ecr.ap-southeast-1.amazonaws.com
```

Simply:

```
go mod init producer
docker build --tag thor:0.0.5 .
```

Or using the custom build script with registry config (login prior to this):

```
(base) welcome@Traianos-MacBook-Pro producer % ./build.sh -n load-tester -t 0.0.1 -p true -d true
Building docker image with options: 
Name: hammer
Tag: 0.0.1
Push: true
building image hammer with tag 0.0.1 and pushing to 605125156525.dkr.ecr.us-east-2.amazonaws.com repo hammer
running build command docker build --tag hammer:0.0.1 .
[+] Building 4.5s (19/19) FINISHED                                                                                                                                                                                                                                                                                 
 => [internal] load build definition from Dockerfile                                                                                                                                                                                                                                                          0.0s
 => => transferring dockerfile: 37B                                                                                                                                                                                                                                                                           0.0s
 => [internal] load .dockerignore                                                                                                                                                                                                                                                                             0.0s
 => => transferring context: 2B                                                                                                                                                                                                                                                                               0.0s
 => resolve image config for docker.io/docker/dockerfile:1                                                                                                                                                                                                                                                    2.6s
 => CACHED docker-image://docker.io/docker/dockerfile:1@sha256:42399d4635eddd7a9b8a24be879d2f9a930d0ed040a61324cfdf59ef1357b3b2                                                                                                                                                                               0.0s
 => [internal] load build definition from Dockerfile                                                                                                                                                                                                                                                          0.0s
 => [internal] load .dockerignore                                                                                                                                                                                                                                                                             0.0s
 => [internal] load metadata for docker.io/library/golang:1.16-alpine                                                                                                                                                                                                                                         1.5s
 => [ 1/10] FROM docker.io/library/golang:1.16-alpine@sha256:41610aabe4ee677934b08685f7ffbeaa89166ed30df9da3f569d1e63789e1992                                                                                                                                                                                 0.0s
 => [internal] load build context                                                                                                                                                                                                                                                                             0.0s
 => => transferring context: 84B                                                                                                                                                                                                                                                                              0.0s
 => CACHED [ 2/10] WORKDIR /app                                                                                                                                                                                                                                                                               0.0s
 => CACHED [ 3/10] COPY go.mod ./                                                                                                                                                                                                                                                                             0.0s
 => CACHED [ 4/10] COPY go.sum ./                                                                                                                                                                                                                                                                             0.0s
 => CACHED [ 5/10] RUN go mod download                                                                                                                                                                                                                                                                        0.0s
 => CACHED [ 6/10] COPY *.go ./                                                                                                                                                                                                                                                                               0.0s
 => CACHED [ 7/10] RUN go build -o /producer                                                                                                                                                                                                                                                                  0.0s
 => CACHED [ 8/10] RUN mkdir -p /datastore                                                                                                                                                                                                                                                                    0.0s
 => CACHED [ 9/10] RUN mkdir -p /processed                                                                                                                                                                                                                                                                    0.0s
 => CACHED [10/10] RUN mkdir -p /applogs/                                                                                                                                                                                                                                                                     0.0s
 => exporting to image                                                                                                                                                                                                                                                                                        0.0s
 => => exporting layers                                                                                                                                                                                                                                                                                       0.0s
 => => writing image sha256:86883e6db410e787713bd09601461e7cd28aa3978c0ea01998d022c1940605d9                                                                                                                                                                                                                  0.0s
 => => naming to docker.io/library/hammer:0.0.1                                                                                                                                                                                                                                                               0.0s

Use 'docker scan' to run Snyk tests against images to find vulnerabilities and learn how to fix them

pushing docker image: docker push 605125156525.dkr.ecr.us-east-2.amazonaws.com/hammer:0.0.1The push refers to repository [605125156525.dkr.ecr.us-east-2.amazonaws.com/hammer] 5d174fa0eb70: Preparing 549a79b81723: Preparing 16840b8df015: Preparing c4027ae51b81: Preparing 903fa0d61f6c: Preparing 88971a11698b: Preparing 00a4fb3dbb29: Preparing 35c63a883fd2: Preparing aa147b7c1d1f: Preparing 19c4d4cefc09: Preparing 46e96c819e17: Preparing b6f786c730a9: Preparing 63a6bdb95b08: Preparing 8d3ac3489996: Preparing 35c63a883fd2: Waiting aa147b7c1d1f: Waiting 88971a11698b: Waiting 00a4fb3dbb29: Waiting 19c4d4cefc09: Waiting 8d3ac3489996: Waiting b6f786c730a9: Waiting 46e96c819e17: Waiting 16840b8df015: Layer already exists 549a79b81723: Layer already exists c4027ae51b81: Layer already exists 903fa0d61f6c: Layer already exists 5d174fa0eb70: Layer already exists 88971a11698b: Layer already exists 00a4fb3dbb29: Layer already exists 35c63a883fd2: Layer already exists aa147b7c1d1f: Layer already exists 19c4d4cefc09: Layer already exists 46e96c819e17: Layer already exists b6f786c730a9: Layer already exists 8d3ac3489996: Layer already exists 63a6bdb95b08: Layer already exists 0.0.1: digest: sha256:957908eec42149b1a6d7aa9dff955e3a71d1abc20a9d3720f2734946f0e640a5 size: 3240

REPOSITORY                                            TAG       IMAGE ID       CREATED        SIZE
REDACTED.dkr.ecr.us-east-2.amazonaws.com/hammer   0.0.1     86883e6db410   3 hours ago    544MB
```

- Deploy to kubnernetes:

```
(base) welcome@Traianos-MacBook-Pro producer % kubectl apply -f producer-manifest.yaml
persistentvolumeclaim/datastore-claim created
persistentvolumeclaim/processed-claim created
persistentvolumeclaim/applogs-claim created
deployment.apps/producer created
```

# Access to the registry

For AWS ECR:

- Get the login credential

```
aws ecr get-login-password --region us-east-2 
```

Create the secret using the docker-password obtain from the last step:

```
kubectl create secret docker-registry ragnarok \
    --namespace ragnarok \
    --docker-server=REDACTED.dkr.ecr.us-east-2.amazonaws.com \
    --docker-username=AWS \
    --docker-password="REDACTED"
```

# Technical Challenges

```
[] Scheme for concurrent reading of a directory of input files
[] Too many file handles open connecting to Kafka
[] Matching concurrency to System Resources
[] How to customise payload to the target api (Allow for API customisation for loading)
```

# Troubleshooting

1)

```
incorrect message format (not readable json)unexpected end of JSON input
Error Count:  81
wrote:   to topic  metrics
panic: worker4could not create message payload for API: open ../output-api/37a80a7d-cb04-4a44-b3f7-5e39d40d93634: too many open files

goroutine 9 [running]:
main.consume_payload_data(0x13bd860, 0xc00001a0a8, 0x135be59, 0x8, 0x4, 0x0, 0x0)
        /Users/welcome/Desktop/STUFF/umbrellacorp/lolcorp-load-generator/microservices/consumer.go:163 +0xa3f
main.worker(0x4, 0xc0001d0000, 0xc0001e4000)
        /Users/welcome/Desktop/STUFF/umbrellacorp/lolcorp-load-generator/microservices/consumer.go:188 +0x85
created by main.main
        /Users/welcome/Desktop/STUFF/umbrellacorp/lolcorp-load-generator/microservices/consumer.go:213 +0x9f
```

2)

```
error  open ../datastore/123: no such file or directory skipping over  123
panic: could not write message dial tcp 127.0.0.1:9092: socket: too many open filesto topicmessages

goroutine 33 [running]:
main.produce(0x0, 0x0, 0x13440c0, 0xc0000a2000, 0x12ecec9, 0x8)
        /Users/welcome/Desktop/STUFF/umbrellacorp/lolcorp-load-generator/microservices/producer.go:96 +0x465
main.process_input_data(0xf, 0x79, 0xc00062ff01, 0x2)
        /Users/welcome/Desktop/STUFF/umbrellacorp/lolcorp-load-generator/microservices/producer.go:61 +0x2c5
main.worker(0xf, 0xc000184000, 0xc000188000)
        /Users/welcome/Desktop/STUFF/umbrellacorp/lolcorp-load-generator/microservices/producer.go:118 +0x4f
created by main.main
        /Users/welcome/Desktop/STUFF/umbrellacorp/lolcorp-load-generator/microservices/producer.go:143 +0x9f
```


3)

```
(base) welcome@Traianos-MacBook-Pro microservices % ./producer          
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x18 pc=0x1266e3a]

goroutine 41 [running]:
main.process_input_data(0x3, 0x4, 0x1)
        /Users/welcome/Desktop/STUFF/umbrellacorp/lolcorp-load-generator/microservices/producer.go:91 +0x37a
main.worker(0x3, 0xc0002a4000, 0xc0002a40b0)
        /Users/welcome/Desktop/STUFF/umbrellacorp/lolcorp-load-generator/microservices/producer.go:166 +0x5d
created by main.main
        /Users/welcome/Desktop/STUFF/umbrellacorp/lolcorp-load-generator/microservices/producer.go:254 +0xff
(base) welcome@Traianos-MacBook-Pro microservices % ./producer
```



4)

```
fatal error: concurrent map writes
fatal error: concurrent map writes

goroutine 14 [running]:
        goroutine running on other thread; stack unavailable
created by main.main
        /Users/welcome/Desktop/STUFF/umbrellacorp/lolcorp-load-generator/microservices/producer.go:288 +0xff

```

5) 


```
fatal error: concurrent map writes


```


Technical Notes
===============


1. Internal DNS Endpoints

- By convention kubernetes  will assign the following fqdns to the services in the relevant namespaces:

```
1. Kafka broker URL                    : kafka.kafka.svc.cluster.local:9092
2. Producer Service (metrics port)     : producer-service.ragnarok.svc.cluster.local:80
3. Consumer Service (metrics port)     : consumer-service.ragnarok.svc.cluster.local:80
4. Sink Service (API and metrics port) : sink-service.ragnarok.svc.cluster.local:80
5. NFS Service                         : nfs-service.ragnarok.svc.cluster.local: 2049 
```


2. Installing kafka on k8s:

```
(base) welcome@Traianos-MacBook-Pro kafka % helm install kafka incubator/kafka -f values.yaml --namespace kafka
WARNING: This chart is deprecated
W1222 13:48:56.124818    2637 warnings.go:70] policy/v1beta1 PodDisruptionBudget is deprecated in v1.21+, unavailable in v1.25+; use policy/v1 PodDisruptionBudget
W1222 13:48:56.651731    2637 warnings.go:70] policy/v1beta1 PodDisruptionBudget is deprecated in v1.21+, unavailable in v1.25+; use policy/v1 PodDisruptionBudget
NAME: kafka
LAST DEPLOYED: Wed Dec 22 13:48:55 2021
NAMESPACE: kafka
STATUS: deployed
REVISION: 1
NOTES:
### Connecting to Kafka from inside Kubernetes

You can connect to Kafka by running a simple pod in the K8s cluster like this with a configuration like this:

  apiVersion: v1
  kind: Pod
  metadata:
    name: testclient
    namespace: kafka
  spec:
    containers:
    - name: kafka
      image: confluentinc/cp-kafka:5.0.1
      command:
        - sh
        - -c
        - "exec tail -f /dev/null"

Once you have the testclient pod above running, you can list all kafka
topics with:

  kubectl -n kafka exec testclient -- ./bin/kafka-topics.sh --zookeeper kafka-zookeeper:2181 --list

To create a new topic:

  kubectl -n kafka exec testclient -- ./bin/kafka-topics.sh --zookeeper kafka-zookeeper:2181 --topic test1 --create --partitions 1 --replication-factor 1

To listen for messages on a topic:

  kubectl -n kafka exec -ti testclient -- ./bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic test1 --from-beginning

To stop the listener session above press: Ctrl+C

To start an interactive message producer session:
  kubectl -n kafka exec -ti testclient -- ./bin/kafka-console-producer.sh --broker-list kafka-headless:9092 --topic test1

To create a message in the above session, simply type the message and press "enter"
To end the producer session try: Ctrl+C

If you specify "zookeeper.connect" in configurationOverrides, please replace "kafka-zookeeper:2181" with the value of "zookeeper.connect", or you will get error.
(base) welcome@Traianos-MacBook-Pro kafka % 
```
