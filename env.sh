#!/bin/bash
#Kubernetes cluster settings for AKS or EKS
export TOKEN="$1"
export AWS_ID="traiano.welcome.contractor"
export ENV="dev"
export target_cloud="aws"
export AWS_DEPLOY_REGION="ap-southeast-1"
export AWS_CLUSTER_NAME="ragnarok-eks-mjollner-poc"
export AWS_ACCOUNT_NUMBER="168393062562"
export AKS_RESOURCE_GROUP="anvil-mjollner-rg"
export AKS_CLUSTER_NAME="ragnarok-aks-mjollner-poc"
export PUBLIC_IP_SKU="Standard"
export IP_ALLOCATION_METHOD="static"
export DEPLOYMENT_NAME="${AKS_CLUSTER_NAME}_SP"

function assumeRoleGetToken() {

    unset AWS_SECRET_ACCESS_KEY AWS_SESSION_TOKEN AWS_ACCESS_KEY_ID
    OUTPUT=`aws sts assume-role --role-arn arn:aws:iam::168393062562:role/admin --role-session-name $ENV --serial-number arn:aws:iam::929981421241:mfa/$AWS_ID --duration-seconds 3600 --token-code $TOKEN`
    echo $OUTPUT
    export AWS_ACCESS_KEY_ID="`echo $OUTPUT | jq -r .Credentials.AccessKeyId`"
    export AWS_SECRET_ACCESS_KEY="`echo $OUTPUT | jq -r .Credentials.SecretAccessKey`"
    export AWS_SESSION_TOKEN="`echo $OUTPUT | jq -r .Credentials.SessionToken`"

    printf "getting calller identitity \n"
    aws sts get-caller-identity

    printf "listing some resources ..."
    aws ec2 describe-instances --region ap-southeast-1
}

assumeRoleGetToken

