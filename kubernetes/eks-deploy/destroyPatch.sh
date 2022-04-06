#!/bin/bash
#cleanup helper script to discover dependencies and cleanup an unclean terraform destroy
#https://aws.amazon.com/premiumsupport/knowledge-center/troubleshoot-dependency-error-delete-vpc/
#this is needed for now because terraform destroy of the EKS module not  clean.

vpc="vpc-07124ed214196db08"
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
    echo  "delete security groups> $i" - $(aws ec2 delete-security-group --group-id $i --region $region)
  done
}

function remove_network_interfaces () {
  for i in `aws ec2 describe-network-interfaces --region $region --filters "Name=vpc-id,Values=$vpc" |  jq -r '."NetworkInterfaces"' | jq -r '.[] | .NetworkInterfaceId'`
    do echo "delete nics> $i" - $(aws ec2 delete-network-interface --network-interface-id $i --region $region)
  done
}

function remove_nat_gw () {
 for i in `aws ec2 describe-nat-gateways --region $region --filter "Name=vpc-id,Values=$vpc" | jq -r '."NatGateways"| .[]' | jq -r '.NatGatewayId'`
 do 
   echo "delete network gw> $i" #- $(aws ec2 delete-nat-gateway --nat-gateway-id $i  --region $region)
 done
}

function remove_load_balancers () {
   for i in `aws elb describe-load-balancers --region $region --query "LoadBalancerDescriptions[?VPCId=='$vpc']|[].LoadBalancerName" | jq -r '.[]'`
   do 
     printf "will delete elb> $i"
     OUT=$(aws elb delete-load-balancer --load-balancer-name $i --region $region)
     printf "\nDELETED LB> $OUT\n"
   done
 
  #aws elbv2 describe-load-balancers
  for i in `aws elbv2 describe-load-balancers --region $region --query "LoadBalancers[?VpcId=='$vpc']|[].LoadBalancerArn" | jq -r '.[]'`
  do
    printf "will delete elb> $i"
    OUT=$(aws elbv2 delete-load-balancer --load-balancer-arn $i --region $region)
    printf "\nDELETED ELB> $OUT\n"
  done
}

function delete_ec2_instances () {
  echo "aws ec2 describe-instances --region $region --filters "Name=vpc-id,Values=$vpc" --query "Reservations[].Instances[].InstanceId" | jq -r '.[]'"
  for i in `aws ec2 describe-instances --region $region --filters "Name=vpc-id,Values=$vpc" --query "Reservations[].Instances[].InstanceId" | jq -r '.[]'`
  do
   echo "delete ec2 instances> $i"
   echo "aws ec2 terminate-instances --instance-ids $i --region $region"
  done
}

function delete_asgs () {
  for i in `aws autoscaling describe-auto-scaling-groups --region $region --filters "Name=vpc-id,Values=$vpc" | jq -r '.[]|.[]|.AutoScalingGroupName'`
    do echo "delete autoscaling groups> $i"
    #- $(aws autoscaling delete-auto-scaling-group --auto-scaling-group-name $i --region $region)
  done
}

function delete_launch_configurations () {
  for i in `aws autoscaling describe-launch-configurations --region $region | jq '."LaunchConfigurations"|.[]."LaunchConfigurationName"'|jq -r ''`
  do
    echo "deleting launch configuration> $i" 
    #aws autoscaling delete-launch-configuration --launch-configuration-name $i --filters "Name=vpc-id,Values=$vpc" --region $region
  done
}

function fixup_state () {
 echo terraform state rm module.eks.kubernetes_config_map.aws_auth
 terraform state rm module.eks.kubernetes_config_map.aws_auth
}

function delete_vpc () {
    OUT=$(aws ec2 delete-vpc --vpc-id $vpc)
    printf "$OUT\n"
}

##delete load balancer dependency
#list_all_dependents
#remove_nat_gw
#delete_launch_configurations
#delete_ec2_instances
delete_asgs
#remove_security_groups
#remove_load_balancers
#remove_network_interfaces
#fixup_state
#delete_vpc
#list_all_dependents
