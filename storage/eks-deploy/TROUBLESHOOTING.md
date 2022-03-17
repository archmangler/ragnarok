
- Check the eks s3 service account is present:

```
(base) welcome@Traianos-MacBook-Pro eks-deploy % kubectl get serviceaccounts -n ragnarok
NAME               SECRETS   AGE
default            1         30m
eks-s3-access      1         12m
internal-kubectl   1         25m
```

- Deployment should looks like this:

```
(base) welcome@Traianos-MacBook-Pro eks-deploy % ./deploy.sh 
eksctl utils associate-iam-oidc-provider --cluster=ragnarok-eks-mjollner-poc
2022-03-17 11:00:41 [ℹ]  eksctl version 0.77.0
2022-03-17 11:00:41 [ℹ]  using region ap-southeast-1
2022-03-17 11:00:42 [ℹ]  IAM Open ID Connect provider is already associated with cluster "ragnarok-eks-mjollner-poc" in "ap-southeast-1"
eksctl delete iamserviceaccount --cluster=ragnarok-eks-mjollner-poc --name=eks-s3-access --namespace=ragnarok
2022-03-17 11:00:43 [ℹ]  eksctl version 0.77.0
2022-03-17 11:00:43 [ℹ]  using region ap-southeast-1
2022-03-17 11:00:45 [ℹ]  1 iamserviceaccount (ragnarok/eks-s3-access) was included (based on the include/exclude rules)
2022-03-17 11:00:45 [ℹ]  1 task: { delete serviceaccount "ragnarok/eks-s3-access" }
2022-03-17 11:00:46 [ℹ]  serviceaccount "ragnarok/eks-s3-access" was already deleted
eksctl create iamserviceaccount --cluster=ragnarok-eks-mjollner-poc --name=eks-s3-access --namespace=ragnarok --attach-policy-arn=arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess --approve --override-existing-serviceaccounts
2022-03-17 11:00:48 [ℹ]  eksctl version 0.77.0
2022-03-17 11:00:48 [ℹ]  using region ap-southeast-1
2022-03-17 11:00:49 [ℹ]  3 existing iamserviceaccount(s) (kube-system/aws-load-balancer-controller,nginx-ingress/eksingress,ragnarok/iam-s3) will be excluded
2022-03-17 11:00:49 [ℹ]  1 iamserviceaccount (ragnarok/eks-s3-access) was included (based on the include/exclude rules)
2022-03-17 11:00:49 [!]  metadata of serviceaccounts that exist in Kubernetes will be updated, as --override-existing-serviceaccounts was set
2022-03-17 11:00:49 [ℹ]  1 task: { 
    2 sequential sub-tasks: { 
        create IAM role for serviceaccount "ragnarok/eks-s3-access",
        create serviceaccount "ragnarok/eks-s3-access",
    } }2022-03-17 11:00:49 [ℹ]  building iamserviceaccount stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-ragnarok-eks-s3-access"
2022-03-17 11:00:49 [ℹ]  deploying stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-ragnarok-eks-s3-access"
2022-03-17 11:00:49 [ℹ]  waiting for CloudFormation stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-ragnarok-eks-s3-access"
2022-03-17 11:01:06 [ℹ]  waiting for CloudFormation stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-ragnarok-eks-s3-access"
2022-03-17 11:01:22 [ℹ]  waiting for CloudFormation stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-ragnarok-eks-s3-access"
2022-03-17 11:01:42 [ℹ]  waiting for CloudFormation stack "eksctl-ragnarok-eks-mjollner-poc-addon-iamserviceaccount-ragnarok-eks-s3-access"
2022-03-17 11:01:43 [ℹ]  created serviceaccount "ragnarok/eks-s3-access"
(base) welcome@Traianos-MacBook-Pro eks-deploy % 
```
