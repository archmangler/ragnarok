#!/bin/bash
#update/refresh settings from values.yaml
helm upgrade pulsar kafkaesque/pulsar --namespace pulsar -f values.yaml
