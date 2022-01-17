variable "appId" {
  description = "Azure Kubernetes Service Cluster service principal"
}

variable "password" {
  description = "Azure Kubernetes Service Cluster password"
}

variable "project_prefix" {
  description = "unique project prefix for resource naming strings"
  default     = "anvil-mjolner"
}

variable "np001_node_count" {
  description = "maximum number of nodes in the default node pool"
  default     = 6
}

variable "np001_node_size" {
  description = "node server sizes in the default node pool"
  default       = "Standard_D4_v4"
}
variable "np001_node_disk_size" {
  description = "size of local disk on default nodepool node server"
  default       = "50"
}
