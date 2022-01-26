#!/bin/bash
helm uninstall kafka incubator/kafka --namespace kafka
watch kubectl get pods -n kafka
