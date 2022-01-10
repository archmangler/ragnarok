#!/bin/bash
#update/refresh settings from values.yaml
helm uninstall kafka incubator/kafka --namespace kafka
