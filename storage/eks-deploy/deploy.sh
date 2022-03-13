#!/bin/bash

function s3_deploy () {
 printf "initializing S3 terraform state ..."
 terraform init
 terraform plan -out terraform.plan
 terraform apply terraform.plan
}

s3_deploy
