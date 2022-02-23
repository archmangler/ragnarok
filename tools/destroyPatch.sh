#!/bin/bash
#cleanup helper script to discover dependencies and cleanup an unclean terraform destroy
#https://aws.amazon.com/premiumsupport/knowledge-center/troubleshoot-dependency-error-delete-vpc/
#this is needed for now because terraform destroy of the EKS module not  clean.

vpc="vpc-00e943247bc428623"
region="ap-southeast-1"

function list_all_dependents () {
 aws ec2 describe-internet-gateways --region $region --filters 'Name=attachment.vpc-id,Values='$vpc | grep InternetGatewayId
 aws ec2 describe-subnets --region $region --filters 'Name=vpc-id,Values='$vpc | grep SubnetId
 aws ec2 describe-route-tables --region $region --filters 'Name=vpc-id,Values='$vpc | grep RouteTableId
 aws ec2 describe-network-acls --region $region --filters 'Name=vpc-id,Values='$vpc | grep NetworkAclId
 aws ec2 describe-vpc-peering-connections --region $region --filters 'Name=requester-vpc-info.vpc-id,Values='$vpc | grep VpcPeeringConnectionId
 aws ec2 describe-vpc-endpoints --region $region --filters 'Name=vpc-id,Values='$vpc | grep VpcEndpointId
 aws ec2 describe-nat-gateways --region $region --filter 'Name=vpc-id,Values='$vpc | grep NatGatewayId
 aws ec2 describe-security-groups --region $region  --region $region --filters 'Name=vpc-id,Values='$vpc | grep GroupId
 aws ec2 describe-instances  --region $region --filters 'Name=vpc-id,Values='$vpc | grep InstanceId
 aws ec2 describe-vpn-connections --region $region --filters 'Name=vpc-id,Values='$vpc | grep VpnConnectionId
 aws ec2 describe-vpn-gateways  --region $region --filters 'Name=attachment.vpc-id,Values='$vpc | grep VpnGatewayId
 aws ec2 describe-network-interfaces --region $region --filters 'Name=vpc-id,Values='$vpc | grep NetworkInterfaceId
}

function remove_security_groups () {
  for i in `aws ec2 describe-security-groups --region $region --filters "Name=vpc-id,Values=$vpc" | jq -r ".SecurityGroups | .[]|.GroupId"`
  do
    echo  "delete> $i" - $(aws ec2 delete-security-group --group-id $i --region $region)
  done
}

function remove_network_interfaces () {
  for i in `aws ec2 describe-network-interfaces --region $region |  jq -r '."NetworkInterfaces"' | jq -r '.[] | .NetworkInterfaceId'`
    do echo "$i" - $(aws ec2 delete-network-interface --network-interface-id $i --region $region)
  done
}

function remove_load_balancers () {
  #aws elb (v1) describe-load-balancers
   for i in `aws elb describe-load-balancers --region $region --query "LoadBalancerDescriptions[?VPCId=='$vpc']|[].LoadBalancerName" | jq -r '.[]'`
   do 
     printf "will delete> $i"
     OUT=$(aws elb delete-load-balancer --load-balancer-name $i --region $region)
     printf "\nDELETED> $OUT\n"
   done
 
  #aws elbv2 describe-load-balancers
  for i in `aws elbv2 describe-load-balancers --region $region --query "LoadBalancers[?VpcId=='$vpc']|[].LoadBalancerArn" | jq -r '.[]'`
  do
    printf "will delete> $i"
    OUT=$(aws elbv2 delete-load-balancer --load-balancer-arn $i --region $region)
    printf "\nDELETED> $OUT\n"
  done
}

function fixup_state () {
 echo terraform state rm module.eks.kubernetes_config_map.aws_auth
 terraform state rm module.eks.kubernetes_config_map.aws_auth
}

#delete load balancer dependency
list_all_dependents
remove_security_groups
remove_load_balancers
remove_network_interfaces
fixup_state
list_all_dependents
