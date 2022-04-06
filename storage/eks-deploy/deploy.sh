#!/bin/bash
namespace="ragnarok"
POLICY_ARN="arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"

function s3_deploy () {
 printf "initializing S3 terraform state ..."
 terraform init
 terraform plan -out terraform.plan
 terraform apply terraform.plan
}

function create_eks_storage_access() {
  echo "eksctl utils associate-iam-oidc-provider --cluster=$AWS_CLUSTER_NAME"
  eksctl utils associate-iam-oidc-provider --cluster=$AWS_CLUSTER_NAME 
  echo "eksctl delete iamserviceaccount --cluster=$AWS_CLUSTER_NAME --name=eks-s3-access --namespace=$namespace"
  eksctl delete iamserviceaccount --cluster=$AWS_CLUSTER_NAME --name=eks-s3-access --namespace=$namespace
  echo "eksctl create iamserviceaccount --cluster=$AWS_CLUSTER_NAME --name=eks-s3-access --namespace=$namespace --attach-policy-arn="$POLICY_ARN" --approve --override-existing-serviceaccounts"
  eksctl create iamserviceaccount --cluster=$AWS_CLUSTER_NAME --name=eks-s3-access --namespace=$namespace --attach-policy-arn="$POLICY_ARN" --approve --override-existing-serviceaccounts
}

s3_deploy
create_eks_storage_access
