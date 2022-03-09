module "eks" {
  source          = "terraform-aws-modules/eks/aws"
  version         = "17.24.0"
  cluster_name    = local.cluster_name
  cluster_version = "1.20"
  subnets         = module.vpc.private_subnets

  vpc_id = module.vpc.vpc_id

  workers_group_defaults = {
    root_volume_type = "gp2"
  }

  worker_groups = [
    {
      name                          = "np001"
      instance_type                 = "t2.medium"
      additional_userdata           = "echo basic management functions"
      additional_security_group_ids = [aws_security_group.worker_group_mgmt_one.id]
      asg_desired_capacity          = 6
      asg_max_size                  = 15
      labels = {
       nodegroup = "np001"
      }
    },
    {
      name                          = "np002"
      instance_type                 = "t2.medium"
      additional_userdata           = "echo basic support functions"
      additional_security_group_ids = [aws_security_group.worker_group_mgmt_two.id]
      asg_desired_capacity          = 6
      asg_max_size                  = 15
    },
    {
      name                          = "np003"
      instance_type                 = "t2.medium"
      additional_userdata           = "echo messaging bus pool"
      additional_security_group_ids = [aws_security_group.worker_group_mgmt_two.id]
      asg_desired_capacity          = 6
      asg_max_size                  = 15
    },
    {
      name                          = "np004"
      instance_type                 = "t2.medium"
      additional_userdata           = "echo producer pool"
      additional_security_group_ids = [aws_security_group.worker_group_mgmt_two.id]
      asg_desired_capacity          = 6
      asg_max_size                  = 15
    },
    {
      name                          = "np005"
      instance_type                 = "t2.medium"
      additional_userdata           = "echo consumer pool"
      additional_security_group_ids = [aws_security_group.worker_group_mgmt_two.id]
      asg_desired_capacity          = 6
      asg_max_size                  = 15
    },
  ]
}

data "aws_eks_cluster" "cluster" {
  name = module.eks.cluster_id
}

data "aws_eks_cluster_auth" "cluster" {
  name = module.eks.cluster_id
}
