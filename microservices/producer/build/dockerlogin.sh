#!/bin/bash
printf "logging into docker registry\n"
aws ecr get-login-password --region ap-southeast-1 | docker login --username AWS --password-stdin 605125156525.dkr.ecr.ap-southeast-1.amazonaws.com
