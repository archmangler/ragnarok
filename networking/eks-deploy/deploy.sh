#!/bin/bash
#Deploy of AWS ALB Ingress controller for EKS
#See: 
# https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html#lbc-install-controller
# https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html
#  https://docs.aws.amazon.com/eks/latest/userguide/alb-ingress.html

#WARNING: source these from env ...
DEPLOY_REGION="ap-southeast-1"
CLUSTER_NAME="ragnarok-eks-mjollner-poc"
AWS_ACCOUNT_NUMBER="524513049339"

#curl -o iam_policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.3.1/docs/install/iam_policy.json

eksctl utils associate-iam-oidc-provider \
  --region $DEPLOY_REGION \
  --cluster ${CLUSTER_NAME} \
   --approve

aws iam create-policy \
    --policy-name AWSLoadBalancerControllerIAMPolicy \
    --policy-document file://iam_policy.json

eksctl create iamserviceaccount \
  --cluster=${CLUSTER_NAME}  \
  --namespace=kube-system \
  --name=aws-load-balancer-controller \
  --attach-policy-arn=arn:aws:iam::${AWS_ACCOUNT_NUMBER}:policy/AWSLoadBalancerControllerIAMPolicy \
  --override-existing-serviceaccounts \
  --approve \
  --region ${DEPLOY_REGION} 

helm repo add eks https://aws.github.io/eks-charts

helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=${CLUSTER_NAME} \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller

#upgrade sequence
#See: 
# https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html#lbc-install-controller
# https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html
#  https://docs.aws.amazon.com/eks/latest/userguide/alb-ingress.html

kubectl apply -k "github.com/aws/eks-charts/stable/aws-load-balancer-controller/crds?ref=master"
helm upgrade aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=ragnarok-eks-mjollner-poc \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller

for i in `seq 1 10`
do
  kubectl get deployment -n kube-system alb-ingress-controller
  printf "waiting ... $OUT"
  sleep 5
done

