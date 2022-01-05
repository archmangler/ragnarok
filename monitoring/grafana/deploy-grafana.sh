#!/bin/bash
printf "Deploying grafana ...\n"
for i in grafana-deployment.yaml service.yaml grafana-datasource-config.yaml;do echo "$i" - $(kubectl apply -f $i);done
