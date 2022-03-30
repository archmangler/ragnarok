module "eks" {
  source          = "terraform-aws-modules/eks/aws"
  version         = "18.9.0"
  cluster_name    = local.cluster_name
  cluster_version = "1.21"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets


  cluster_addons = {
    coredns = {
      resolve_conflicts = "OVERWRITE"
    }
    kube-proxy = {}
    vpc-cni = {
      resolve_conflicts = "OVERWRITE"
    }
  }


  # EKS Managed Node Group(s)
  eks_managed_node_group_defaults = {
    ami_type               = "AL2_x86_64"
    disk_size              = 50
    instance_types         = ["t2.medium"]
    vpc_security_group_ids = [aws_security_group.worker_group_mgmt_one.id, aws_security_group.worker_group_mgmt_two.id]
  }

  eks_managed_node_groups = {
    np001 = {
      min_size       = 9
      max_size       = 18
      desired_size   = 12
      instance_types = ["t2.xlarge"]
      disk_size      = 100
      capacity_type  = "ON_DEMAND"
      labels = {
        Environment = "poc"
        function    = "management"
      }
    }
    np002 = {
      min_size       = 6
      max_size       = 18
      desired_size   = 12
      instance_types = ["t2.medium"]
      disk_size      = 50
      capacity_type  = "ON_DEMAND"
      labels = {
        Environment = "poc"
        function    = "producers"
      }
    }
    np003 = {
      min_size       = 6
      max_size       = 18
      desired_size   = 12
      instance_types = ["t2.medium"]
      disk_size      = 50
      capacity_type  = "ON_DEMAND"
      labels = {
        Environment = "poc"
        function    = "pulsar"
      }
    }
    np004 = {
      min_size       = 6
      max_size       = 18
      desired_size   = 12
      instance_types = ["t2.medium"]
      disk_size      = 50
      capacity_type  = "ON_DEMAND"
      labels = {
        Environment = "poc"
        function    = "consumers"
      }
    }
    np005 = {
      min_size       = 6
      max_size       = 18
      desired_size   = 12
      disk_size      = 50
      instance_types = ["t2.medium"]
      capacity_type  = "ON_DEMAND"
      labels = {
        Environment = "poc"
        function    = "storage"
      }
    }

  }
}

data "aws_eks_cluster" "cluster" {
  name = module.eks.cluster_id
}

data "aws_eks_cluster_auth" "cluster" {
  name = module.eks.cluster_id
}
