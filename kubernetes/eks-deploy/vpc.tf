variable "region" {
  default     = "ap-southeast-1"
  description = "AWS region"
}

provider "aws" {
  region = var.region
}

data "aws_availability_zones" "available" {}

locals {
  cluster_name = "ragnarok-eks-mjollner-poc" // Alternartively: cluster_name = "ragnarok-eks-${random_string.suffix.result}"
}

resource "random_string" "suffix" {
  length  = 8
  special = false
}

variable "private_subnets" {
  type = list(any)
  default = [
    "192.168.2.0/23",
    "192.168.4.0/23",
    "192.168.6.0/23"
  ]
}

variable "public_subnets" {
  type = list(any)
  default = [
    "192.168.8.0/23",
    "192.168.10.0/23",
    "192.168.12.0/23"
  ]
}

variable "cidr" {
  type    = string
  default = "192.168.0.0/16"
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "3.2.0"

  name                 = "ragnarok"
  cidr                 = var.cidr
  azs                  = data.aws_availability_zones.available.names
  private_subnets      = var.private_subnets
  public_subnets       = var.public_subnets
  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true

  tags = {
    "kubernetes.io/cluster/${local.cluster_name}" = "shared"
  }

  public_subnet_tags = {
    "kubernetes.io/cluster/${local.cluster_name}" = "shared"
    "kubernetes.io/role/elb"                      = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/cluster/${local.cluster_name}" = "shared"
    "kubernetes.io/role/internal-elb"             = "1"
  }
}
