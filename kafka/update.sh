#!/bin/bash
#update/refresh settings from values.yaml
helm upgrade kafka incubator/kafka -f values.yaml --namespace kafka
