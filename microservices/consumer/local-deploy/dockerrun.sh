#!/bin/bash
#run the producer via docker
#on your local desktop (assumed Mac osx)

docker run \
 --mount source=output-api,destination=/output-api \
 --mount source=consumerlogs,destination=/applogs \
 -e NUM_JOBS=20 \
 -e NUM_WORKERS=20 \
 -e KAFKA_BROKER1_ADDRESS='192.168.65.2:9092' \
 -e KAFKA_BROKER2_ADDRESS='192.168.65.2:9092' \
 -e KAFKA_BROKER3_ADDRESS='192.168.65.2:9092' \
 -e DATA_OUT_DIRECTORY='/output-api/' \
 -e LOCAL_LOGFILE_PATH='/applogs/consumer.log' \
 -e MESSAGE_TOPIC='messages' \
 -e DEADLETTER_TOPIC='deadLetter' \
 -e METRICS_TOPIC='metrics' \
 605125156525.dkr.ecr.ap-southeast-1.amazonaws.com/load-consumer:0.0.2
