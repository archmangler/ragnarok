# Pulsar installation via helm charts

# Topic Creation and Testing

Once Pulsar is running, use the utility container to create and test the required topics

```
#Access the utility container:
kubectl -n pulsar exec -it  pulsar-toolset-0 -- /bin/bash

#create the tenant:
/pulsar/bin/pulsar-admin tenants create ragnarok

#create namespaces:
/pulsar/bin/pulsar-admin namespaces create ragnarok/transactions

#list namespaces under tenant
root@pulsar-toolset-0:/pulsar# /pulsar/bin/pulsar-admin namespaces list ragnarok
"ragnarok/transactions"

#Create the topics
root@pulsar-toolset-0:/pulsar# /pulsar/bin/pulsar-admin topics create persistent://ragnarok/transactions/requests

#list the topics
root@pulsar-toolset-0:/pulsar# /pulsar/bin/pulsar-admin topics list ragnarok/transactions
"persistent://ragnarok/transactions/requests"

#Consuming from a topic
/pulsar/bin/pulsar-client consume \
    persistent://ragnarok/transactions/requests \
    --num-messages 0 \
    --subscription-name sub001 \
    --subscription-type Shared

#List Tenants, Namespaces, Topics
/pulsar/bin/pulsar-client produce \
persistent://ragnarok/transactions/requests \
--num-produce 2 \
--messages "Hello Pulsar"

root@pulsar-toolset-0:/pulsar# /pulsar/bin/pulsar-client produce \
> persistent://ragnarok/transactions/requests \
> --num-produce 2 \
> --messages "Hello Pulsar"
17:48:42.495 [pulsar-client-io-1-1] INFO  org.apache.pulsar.client.impl.ConnectionPool - [[id: 0x35248b6e, L:/10.244.3.2:59816 - R:pulsar-proxy.pulsar.svc.cluster.local/10.0.124.108:6650]] Connected to server
17:48:42.636 [pulsar-client-io-1-1] INFO  org.apache.pulsar.client.impl.ProducerStatsRecorderImpl - Starting Pulsar producer perf with config: {
  "topicName" : "persistent://ragnarok/transactions/requests",
  "producerName" : null,
  "sendTimeoutMs" : 30000,
  "blockIfQueueFull" : false,
  "maxPendingMessages" : 1000,
  "maxPendingMessagesAcrossPartitions" : 50000,
  "messageRoutingMode" : "RoundRobinPartition",
  "hashingScheme" : "JavaStringHash",
  "cryptoFailureAction" : "FAIL",
  "batchingMaxPublishDelayMicros" : 1000,
  "batchingPartitionSwitchFrequencyByPublishDelay" : 10,
  "batchingMaxMessages" : 1000,
  "batchingMaxBytes" : 131072,
  "batchingEnabled" : true,
  "chunkingEnabled" : false,
  "compressionType" : "NONE",
  "initialSequenceId" : null,
  "autoUpdatePartitions" : true,
  "autoUpdatePartitionsIntervalSeconds" : 60,
  "multiSchema" : true,
  "properties" : { }
}
17:48:42.644 [pulsar-client-io-1-1] INFO  org.apache.pulsar.client.impl.ProducerStatsRecorderImpl - Pulsar client config: {
  "serviceUrl" : "pulsar://pulsar-proxy:6650/",
  "authPluginClassName" : null,
  "operationTimeoutMs" : 30000,
  "statsIntervalSeconds" : 60,
  "numIoThreads" : 1,
  "numListenerThreads" : 1,
  "connectionsPerBroker" : 1,
  "useTcpNoDelay" : true,
  "useTls" : false,
  "tlsTrustCertsFilePath" : "",
  "tlsAllowInsecureConnection" : false,
  "tlsHostnameVerificationEnable" : false,
  "concurrentLookupRequest" : 5000,
  "maxLookupRequest" : 50000,
  "maxLookupRedirects" : 20,
  "maxNumberOfRejectedRequestPerConnection" : 50,
  "keepAliveIntervalSeconds" : 30,
  "connectionTimeoutMs" : 10000,
  "requestTimeoutMs" : 60000,
  "initialBackoffIntervalNanos" : 100000000,
  "maxBackoffIntervalNanos" : 60000000000,
  "listenerName" : null,
  "useKeyStoreTls" : false,
  "sslProvider" : null,
  "tlsTrustStoreType" : "JKS",
  "tlsTrustStorePath" : "",
  "tlsTrustStorePassword" : "",
  "tlsCiphers" : [ ],
  "tlsProtocols" : [ ],
  "proxyServiceUrl" : null,
  "proxyProtocol" : null,
  "enableTransaction" : false
}
17:48:42.668 [pulsar-client-io-1-1] INFO  org.apache.pulsar.client.impl.ConnectionPool - [[id: 0xd3c525e9, L:/10.244.3.2:59818 - R:pulsar-proxy.pulsar.svc.cluster.local/10.0.124.108:6650]] Connected to server
17:48:42.669 [pulsar-client-io-1-1] INFO  org.apache.pulsar.client.impl.ClientCnx - [id: 0xd3c525e9, L:/10.244.3.2:59818 - R:pulsar-proxy.pulsar.svc.cluster.local/10.0.124.108:6650] Connected through proxy to target broker at pulsar-broker-0.pulsar-broker.pulsar.svc.cluster.local:6650
17:48:42.681 [pulsar-client-io-1-1] INFO  org.apache.pulsar.client.impl.ProducerImpl - [persistent://ragnarok/transactions/requests] [null] Creating producer on cnx [id: 0xd3c525e9, L:/10.244.3.2:59818 - R:pulsar-proxy.pulsar.svc.cluster.local/10.0.124.108:6650]
17:48:42.707 [pulsar-client-io-1-1] INFO  org.apache.pulsar.client.impl.ProducerImpl - [persistent://ragnarok/transactions/requests] [pulsar-2-3] Created producer on cnx [id: 0xd3c525e9, L:/10.244.3.2:59818 - R:pulsar-proxy.pulsar.svc.cluster.local/10.0.124.108:6650]
17:48:42.742 [main] INFO  com.scurrilous.circe.checksum.Crc32cIntChecksum - SSE4.2 CRC32C provider initialized
17:48:42.844 [main] INFO  org.apache.pulsar.client.impl.PulsarClientImpl - Client closing. URL: pulsar://pulsar-proxy:6650/
17:48:42.854 [pulsar-client-io-1-1] INFO  org.apache.pulsar.client.impl.ProducerImpl - [persistent://ragnarok/transactions/requests] [pulsar-2-3] Closed Producer
17:48:42.860 [pulsar-client-io-1-1] INFO  org.apache.pulsar.client.impl.ClientCnx - [id: 0x35248b6e, L:/10.244.3.2:59816 ! R:pulsar-proxy.pulsar.svc.cluster.local/10.0.124.108:6650] Disconnected
17:48:42.864 [pulsar-client-io-1-1] INFO  org.apache.pulsar.client.impl.ClientCnx - [id: 0xd3c525e9, L:/10.244.3.2:59818 ! R:pulsar-proxy.pulsar.svc.cluster.local/10.0.124.108:6650] Disconnected
17:48:42.868 [main] INFO  org.apache.pulsar.client.cli.PulsarClientTool - 2 messages successfully produced
```

