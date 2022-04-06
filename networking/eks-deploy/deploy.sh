#!/bin/bash
#Deploy of AWS ALB Ingress controller for EKS
#See: 
# https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html#lbc-install-controller
# https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html
# https://docs.aws.amazon.com/eks/latest/userguide/alb-ingress.html

#WARNING: source these from env ...
#Examples:
#export AWS_DEPLOY_REGION="ap-southeast-1"
#export AWS_CLUSTER_NAME="ragnarok-eks-mjollner-poc"
#export AWS_ACCOUNT_NUMBER="524513049339"
AWS_ACCOUNT_NUMBER=$(aws sts get-caller-identity | jq -r .Account | xargs)

#curl -o iam_policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.3.1/docs/install/iam_policy.json

function associate_iam_oidc_provider () {
 OUT=$(eksctl utils associate-iam-oidc-provider \
    --region $AWS_DEPLOY_REGION \
    --cluster ${AWS_CLUSTER_NAME} \
    --approve)
  printf "associate_iam_oidc_provider: $OUT\n"
}

function create_iam_policy () {
  aws iam create-policy \
    --policy-name AWSLoadBalancerControllerIAMPolicy \
    --policy-document file://iam_policy.json
}

function delete_iam_service_account () {
  OUT=$(eksctl delete iamserviceaccount --namespace=kube-system --cluster=${AWS_CLUSTER_NAME} --name=aws-load-balancer-controller --region ${AWS_DEPLOY_REGION} --verbose 9)
  printf "Deleting old serviceaccount: $OUT\n"
  for i in `seq 1 9`
  do
    OUT=$(kubectl get serviceaccounts -n kube-system| egrep "aws-load-balancer-controller")
    printf "checking: $OUT\n"
    sleep 2
  done
}

function create_iam_service_account () {

  echo "create_iam_service_account: eksctl create iamserviceaccount \
    --cluster=${AWS_CLUSTER_NAME}  \
    --namespace=kube-system \
    --name=aws-load-balancer-controller \
    --attach-policy-arn=arn:aws:iam::${AWS_ACCOUNT_NUMBER}:policy/AWSLoadBalancerControllerIAMPolicy \
    --approve \
    --region ${AWS_DEPLOY_REGION}"
#    --override-existing-serviceaccounts \

  OUT=$(eksctl create iamserviceaccount \
    --cluster=${AWS_CLUSTER_NAME}  \
    --namespace=kube-system \
    --name=aws-load-balancer-controller \
    --attach-policy-arn=arn:aws:iam::${AWS_ACCOUNT_NUMBER}:policy/AWSLoadBalancerControllerIAMPolicy \
    --approve \
    --region ${AWS_DEPLOY_REGION} --verbose 5)
#    --override-existing-serviceaccounts \
  printf "OUTPUT:$OUT\n"
}

function uninstall_alb () {
  #first uninstall
  helm uninstall aws-load-balancer-controller eks/aws-load-balancer-controller -n kube-system
  for i in `seq 1 10`
  do
    printf "waiting ..."
    sleep 2
  done
}

function install_alb () {
  helm repo add eks https://aws.github.io/eks-charts
  echo "deploying ALB ngress controller with helm:  helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
    -n kube-system \
    --set clusterName=${AWS_CLUSTER_NAME} \
    --set serviceAccount.create=false \
    --set serviceAccount.name=aws-load-balancer-controller"

  OUT=$(helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
    -n kube-system \
    --set clusterName=${AWS_CLUSTER_NAME} \
    --set serviceAccount.create=false \
    --set serviceAccount.name=aws-load-balancer-controller)
}

function upgrade_alb () {
#upgrade sequence
#See: 
# https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html#lbc-install-controller
# https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html
# https://docs.aws.amazon.com/eks/latest/userguide/alb-ingress.html

  printf "running helm chart upgrade ...\n"

  kubectl apply -k "github.com/aws/eks-charts/stable/aws-load-balancer-controller/crds?ref=master"

  OUT=$(helm upgrade aws-load-balancer-controller eks/aws-load-balancer-controller \
    -n kube-system \
    --set clusterName=${AWS_CLUSTER_NAME} \
    --set serviceAccount.create=false \
    --set serviceAccount.name=aws-load-balancer-controller)
}

function check_alb_deployment () {
  printf "checking ingress controller deployment ...\n"
  for i in `seq 1 10`
  do
    OUT=$(kubectl get deployment aws-load-balancer-controller -n kube-system)
    printf "waiting: ... $OUT\n"
    OUT=$(kubectl get events -n kube-system|egrep "aws-load-balancer-controller")
    printf "Related Events: $OUT \n"
    sleep 5
  done
}

#
associate_iam_oidc_provider
create_iam_policy
delete_iam_service_account
create_iam_service_account
uninstall_alb
install_alb
check_alb_deployment
