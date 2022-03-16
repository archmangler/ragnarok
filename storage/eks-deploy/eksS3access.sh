#!/bin/bash
#Give EKS access to S3 to pull down and process data ...
namespace="ragnarok"
POLICY_ARN="arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
eksctl utils associate-iam-oidc-provider --cluster=$AWS_CLUSTER_NAME --approve
eksctl delete iamserviceaccount --cluster=$AWS_CLUSTER_NAME --name=eks-s3-access --namespace=$namespace --approve
eksctl create iamserviceaccount --cluster=$AWS_CLUSTER_NAME --name=eks-s3-access --namespace=$namespace --attach-policy-arn="$POLICY_ARN" --approve
