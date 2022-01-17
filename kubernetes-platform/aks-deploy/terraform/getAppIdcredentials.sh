#!/bin/bash

#an appId is required for the aks cluster
creds=$(az ad sp create-for-rbac --skip-assignment)

echo $creds

