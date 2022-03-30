#!/bin/bash

TOKEN="$1"
AWS_ID="traiano.welcome.contractor"
ENV="sandbox"

function assumeRoleGetToken() {

    unset AWS_SECRET_ACCESS_KEY AWS_SESSION_TOKEN AWS_ACCESS_KEY_ID
    OUTPUT=`aws sts assume-role --role-arn arn:aws:iam::298482326239:role/admin --role-session-name $ENV --serial-number arn:aws:iam::929981421241:mfa/$AWS_ID --duration-seconds 3600 --token-code $TOKEN`
    echo $OUTPUT
    export AWS_ACCESS_KEY_ID="`echo $OUTPUT | jq -r .Credentials.AccessKeyId`"
    export AWS_SECRET_ACCESS_KEY="`echo $OUTPUT | jq -r .Credentials.SecretAccessKey`"
    export AWS_SESSION_TOKEN="`echo $OUTPUT | jq -r .Credentials.SessionToken`"

    printf "getting caller identity \n"
    aws sts get-caller-identity

    printf "listing some resources ..."
    aws ec2 describe-instances --region ap-southeast-1
}

assumeRoleGetToken
