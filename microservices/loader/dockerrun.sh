#!/bin/bash
#run the loader service via docker

docker run \
 --mount source=datastore,destination=/datastore \
 -e NUM_JOBS=20 \
 -e NUM_WORKERS=20 \
 -e KAFKA_BROKER1_ADDRESS='192.168.65.2:9092' \
 -e KAFKA_BROKER2_ADDRESS='192.168.65.2:9092' \
 -e KAFKA_BROKER3_ADDRESS='192.168.65.2:9092' \
 -e DATA_SOURCE_DIRECTORY='/datastore/' \
 -e DATA_OUT_DIRECTORY='/processed/' \
 -e LOCAL_LOGFILE_PATH='/applogs/producer.log' \
 -e MESSAGE_TOPIC='messages' \
 -e DEADLETTER_TOPIC='deadLetter' \
 -e METRICS_TOPIC='metrics' \
  loader:0.0.1
