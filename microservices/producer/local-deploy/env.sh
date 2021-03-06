#!/bin/bash
#Set environment variables for testing 
NUM_JOBS=20
NUM_WORKERS=20
KAFKA_BROKER1_ADDRESS="192.168.65.2:9092"
KAFKA_BROKER2_ADDRESS="192.168.65.2:9092"
KAFKA_BROKER3_ADDRESS="192.168.65.2:9092"
DATA_SOURCE_DIRECTORY="/datastore/"
DATA_OUT_DIRECTORY="/processed/"
LOCAL_LOGFILE_PATH="/applogs/producer.log"
MESSAGE_TOPIC="messages"
DEADLETTER_TOPIC="deadLetter"
METRICS_TOPIC="metrics"
