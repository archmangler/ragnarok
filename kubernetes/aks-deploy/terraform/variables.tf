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

//node pool 1 - for management services, kakfa/redis and support services
variable "np001_node_count" {
  description = "maximum number of nodes in the default node pool"
  default     = 3
}

variable "np001_node_size" {
  description = "node server sizes in the default node pool"
  default     = "Standard_D4_v4"
}
variable "np001_node_disk_size" {
  description = "size of local disk on default nodepool node server"
  default     = "50"
}

variable "enable_auto_scaling_np001" {
  description = "enable/disable autoscaling for default nodepool"
  default     = true
}

variable "max_node_count_np001" {
  description = "maximum node count for default node pool"
  default     = 3
}

variable "min_node_count_np001" {
  description = "minimum node count for default node pool"
  default     = 1
}

//node pool 2 - for producers and consumers doing the heavy lifting
variable "np002_node_count" {
  description = "maximum number of nodes in the default node pool"
  default     = 3
}

variable "np002_node_size" {
  description = "node server sizes in the default node pool"
  default     = "Standard_D4_v4"
}

variable "np002_node_disk_size" {
  description = "size of local disk on default nodepool node server"
  default     = "50"
}

variable "enable_auto_scaling_np002" {
  description = "enable/disable autoscaling for default nodepool"
  default     = true
}

variable "max_node_count_np002" {
  description = "maximum node count for default node pool"
  default     = 3
}

variable "min_node_count_np002" {
  description = "minimum node count for default node pool"
  default     = 1
}

//node pool 3 - for redis and kafka doing the heavy lifting
variable "np003_node_count" {
  description = "maximum number of nodes in the default node pool"
  default     = 3
}

variable "np003_node_size" {
  description = "node server sizes in the default node pool"
  default     = "Standard_D8_v3"
}

variable "np003_node_disk_size" {
  description = "size of local disk on default nodepool node server"
  default     = "50"
}

variable "enable_auto_scaling_np003" {
  description = "enable/disable autoscaling for default nodepool"
  default     = true
}

variable "max_node_count_np003" {
  description = "maximum node count for this  node pool"
  default     = 9 
}

variable "min_node_count_np003" {
  description = "minimum node count for default node pool"
  default     = 6
}

//node pool 4 - for redis and kafka doing the heavy lifting
variable "np004_node_count" {
  description = "maximum number of nodes in this node pool"
  default     = 9
}

variable "np004_node_size" {
  description = "node server sizes in the default node pool"
  default     = "Standard_D8_v3"
}

variable "np004_node_disk_size" {
  description = "size of local disk on default nodepool node server"
  default     = "50"
}

variable "enable_auto_scaling_np004" {
  description = "enable/disable autoscaling for default nodepool"
  default     = true
}

variable "max_node_count_np004" {
  description = "maximum node count for this node pool"
  default     = 9
}

variable "min_node_count_np004" {
  description = "minimum node count for default node pool"
  default     = 6
}

//node pool 5 - for producers doing the heavy lifting
variable "np005_node_count" {
  description = "maximum number of nodes in the producer node pool"
  default     = 3
}

variable "np005_node_size" {
  description = "node server sizes in the producer node pool"
  default     = "Standard_D8_v3"
}

variable "np005_node_disk_size" {
  description = "size of local disk on producer nodepool node server"
  default     = "50"
}

variable "enable_auto_scaling_np005" {
  description = "enable/disable autoscaling for producer nodepool"
  default     = true
}

variable "max_node_count_np005" {
  description = "maximum node count for default node pool"
  default     = 9
}

variable "min_node_count_np005" {
  description = "minimum node count for default node pool"
  default     = 6
}

