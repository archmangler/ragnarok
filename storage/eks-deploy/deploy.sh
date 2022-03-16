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
  eksctl utils associate-iam-oidc-provider --cluster=$AWS_CLUSTER_NAME --approve
  eksctl delete iamserviceaccount --cluster=$AWS_CLUSTER_NAME --name=eks-s3-access --namespace=$namespace --approve
  eksctl create iamserviceaccount --cluster=$AWS_CLUSTER_NAME --name=eks-s3-access --namespace=$namespace --attach-policy-arn="$POLICY_ARN" --approve
}

create_eks_storage_access
#s3_deploy
