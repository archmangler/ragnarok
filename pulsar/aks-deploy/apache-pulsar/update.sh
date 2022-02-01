#!/bin/bash
#update/refresh settings from values.yaml
#helm get values pulsar-mini --namespace pulsar > pulsar.yaml

helm upgrade pulsar-mini apache/pulsar --namespace pulsar -f pulsar.yaml
