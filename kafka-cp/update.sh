#!/bin/bash
#update/refresh settings from values.yaml
helm upgrade kafka cp-helm-charts -f values.yaml --namespace kafka
