#Experimental: for reference
#null resource to deploy EFS shared storage
#resource "null_resource" "install_efs_csi_driver" {
#  depends_on = [module.eks.aws_eks_cluster]
#  provisioner "local-exec" {
#    command = format("kubectl --kubeconfig %s apply -k 'github.com/kubernetes-sigs/aws-efs-csi-driver/deploy/kubernetes/overlays/stable/?ref=release-1.1'", module.eks.kubeconfig_filename)
#  }
#}
#
#resource "aws_efs_file_system" "datastore" {
#  creation_token = local.cluster_name
#  tags = {
#    Name = "${local.cluster_name}-storage"
#  }
#}
#
#resource "aws_efs_mount_target" "datastore" {
#  count          = length(data.aws_availability_zones.available.names)
#  file_system_id = aws_efs_file_system.datastore.id
#  subnet_id      = module.vpc.private_subnets[count.index]
#}
#
#
#resource "kubernetes_persistent_volume_claim" "datastore" {
#  metadata {
#    name = module.eks.cluster_id
#    namespace = "ragnarok"
#  }
#  spec {
#    access_modes       = ["ReadWriteOnce"]
#    storage_class_name = "efs-sc"
#    resources {
#      requests = {
#        storage = "50Gi"
#      }
#    }
#  }
#}
#
#resource "kubernetes_persistent_volume" "datastore" {
#  metadata {
#    name = module.eks.cluster_id
#  }
#  spec {
#    storage_class_name               = "efs-sc"
#    persistent_volume_reclaim_policy = "Retain"
#    capacity = {
#      storage = "50Gi"
#    }
#    access_modes = ["ReadWriteMany"]
#    persistent_volume_source {
#      nfs {
#        path   = "/"
#        server = aws_efs_file_system.datastore.id
#      }
#    }
#  }
#}

