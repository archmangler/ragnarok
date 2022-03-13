#!/bin/bash
CLUSTER_NAME="ragnarok-eks-mjollner-poc"
for STACK_NAME in $(eksctl get nodegroup --cluster $CLUSTER_NAME -o json | jq -r '.[].StackName')
do
  ROLE_NAME=$(aws cloudformation describe-stack-resources --stack-name $STACK_NAME | jq -r '.StackResources[] | select(.ResourceType=="AWS::IAM::Role") | .PhysicalResourceId')
  aws iam attach-role-policy \
    --role-name $ROLE_NAME \
    --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess
done
