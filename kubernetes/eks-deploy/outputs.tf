output "cluster_id" {
  description = "EKS cluster ID."
  value       = module.eks.cluster_id
}

output "cluster_endpoint" {
  description = "Endpoint for EKS control plane."
  value       = module.eks.cluster_endpoint
}

output "cluster_security_group_id" {
  description = "Security group ids attached to the cluster control plane."
  value       = module.eks.cluster_security_group_id
}

output "region" {
  description = "AWS region"
  value       = var.region
}

output "cluster_name" {
  description = "Kubernetes Cluster Name"
  value       = local.cluster_name
}

output "subnet_zero" {
   description = "public subnet for bastion"
   value =  module.vpc.private_subnets[0]
}

output "security_group_mgmt" {
  description = "sg one"
  value = aws_security_group.all_worker_mgmt.id
}

output "security_group_two" {
  description = "sg two"
  value = aws_security_group.worker_group_mgmt_two.id
}

output "security_group_one" {
  description = "sg one"
  value = aws_security_group.worker_group_mgmt_one.id
}

