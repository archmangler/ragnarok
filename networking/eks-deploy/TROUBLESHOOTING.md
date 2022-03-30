

# EKS service account troubleshooting 


- Creation of an EKS service account for alb should complete as follows:

```
2022-03-22 16:13:56 [▶]  started task: create IAM role for serviceaccount "kube-system/aws-load-balancer-controller"
2022-03-22 16:13:56 [ℹ]  building iamserviceaccount stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-kube-system-aws-load-balancer-controller"
2022-03-22 16:13:56 [▶]  service account location provided: kube-system/aws-node, adding sub condition
2022-03-22 16:13:56 [▶]  CreateStackInput = {
  Capabilities: ["CAPABILITY_IAM"],
  DisableRollback: false,
  StackName: "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-kube-system-aws-load-balancer-controller",
  Tags: [
    {
      Key: "alpha.eksctl.io/cluster-name",
      Value: "ragnarok-eks-mjollner-poc"
    },
    {
      Key: "eksctl.cluster.k8s.io/v1alpha1/cluster-name",
      Value: "ragnarok-eks-mjollner-poc"
    },
    {
      Key: "alpha.eksctl.io/eksctl-version",
      Value: "0.77.0"
    },
    {
      Key: "alpha.eksctl.io/iamserviceaccount-name",
      Value: "kube-system/aws-load-balancer-controller"
    }
  ],
  TemplateBody: "{"AWSTemplateFormatVersion":"2010-09-09","Description":"IAM role for serviceaccount \"kube-system/aws-load-balancer-controller\" [created and managed by eksctl]","Resources":{"Role1":{"Type":"AWS::IAM::Role","Properties":{"AssumeRolePolicyDocument":{"Statement":[{"Action":["sts:AssumeRoleWithWebIdentity"],"Condition":{"StringEquals":{"oidc.eks.ap-southeast-1.amazonaws.com/id/CEBD060A21686580D9F5A04B13A2E5FD:aud":"sts.amazonaws.com","oidc.eks.ap-southeast-1.amazonaws.com/id/CEBD060A21686580D9F5A04B13A2E5FD:sub":"system:serviceaccount:kube-system:aws-load-balancer-controller"}},"Effect":"Allow","Principal":{"Federated":"arn:aws:iam::298482326239:oidc-provider/oidc.eks.ap-southeast-1.amazonaws.com/id/CEBD060A21686580D9F5A04B13A2E5FD"}}],"Version":"2012-10-17"},"ManagedPolicyArns":["arn:aws:iam::298482326239:policy/AWSLoadBalancerControllerIAMPolicy"]}}},"Outputs":{"Role1":{"Value":{"Fn::GetAtt":"Role1.Arn"}}}}"
}
2022-03-22 16:13:56 [ℹ]  deploying stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-kube-system-aws-load-balancer-controller"
2022-03-22 16:13:56 [▶]  start waiting for CloudFormation stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-kube-system-aws-load-balancer-controller"
2022-03-22 16:13:56 [ℹ]  waiting for CloudFormation stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-kube-system-aws-load-balancer-controller"
2022-03-22 16:14:12 [ℹ]  waiting for CloudFormation stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-kube-system-aws-load-balancer-controller"
2022-03-22 16:14:29 [ℹ]  waiting for CloudFormation stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-kube-system-aws-load-balancer-controller"
2022-03-22 16:14:49 [ℹ]  waiting for CloudFormation stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-kube-system-aws-load-balancer-controller"
2022-03-22 16:14:49 [▶]  done after 53.392561868s of waiting for CloudFormation stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-kube-system-aws-load-balancer-controller"
2022-03-22 16:14:49 [▶]  completed task: create IAM role for serviceaccount "kube-system/aws-load-balancer-controller"
2022-03-22 16:14:49 [▶]  started task: create serviceaccount "kube-system/aws-load-balancer-controller"
2022-03-22 16:14:49 [ℹ]  created serviceaccount "kube-system/aws-load-balancer-controller"
2022-03-22 16:14:49 [▶]  completed task: create serviceaccount "kube-system/aws-load-balancer-controller"
2022-03-22 16:14:49 [▶]  completed task: 
    2 sequential sub-tasks: { 
        create IAM role for serviceaccount "kube-system/aws-load-balancer-controller",
        create serviceaccount "kube-system/aws-load-balancer-controller",
    }
```
