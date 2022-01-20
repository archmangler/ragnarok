provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "ragnarok" {
  name     = "${var.project_prefix}-rg"
  location = "southeastasia"

  tags = {
    environment = "poc"
    project     = "trade matching system load testing"
  }
}

resource "azurerm_kubernetes_cluster" "ragnarok" {

  name                = "${var.project_prefix}-benchmarking-cluster"
  location            = azurerm_resource_group.ragnarok.location
  resource_group_name = azurerm_resource_group.ragnarok.name
  dns_prefix          = "${var.project_prefix}-k8s"

  default_node_pool {
    name                = "np001"
    enable_auto_scaling = var.enable_auto_scaling_np001
    max_count           = var.max_node_count_np001
    min_count           = var.min_node_count_np001
    vm_size             = var.np001_node_size //"Standard_D2_v2"
    os_disk_size_gb     = var.np001_node_disk_size
  }

  service_principal {
    client_id     = var.appId
    client_secret = var.password
  }

  role_based_access_control {
    enabled = true
  }

  tags = {
    environment = "poc"
    project     = "trade matching system load testing"
  }
}

resource "azurerm_kubernetes_cluster_node_pool" "workers" {
  name                  = "np002"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.ragnarok.id
  vm_size               = var.np002_node_size
  enable_auto_scaling   = var.enable_auto_scaling_np002
  max_count             = var.max_node_count_np002
  min_count             = var.min_node_count_np002
  os_disk_size_gb       = var.np002_node_disk_size

  tags = {
    environment = "poc"
    project     = "trade matching system load testing"
  }
}

resource "azurerm_kubernetes_cluster_node_pool" "dataplane" {
  name                  = "np003"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.ragnarok.id
  vm_size               = var.np003_node_size
  enable_auto_scaling   = var.enable_auto_scaling_np003
  max_count             = var.max_node_count_np003
  min_count             = var.min_node_count_np003
  os_disk_size_gb       = var.np003_node_disk_size

  tags = {
    environment = "poc"
    project     = "trade matching system load testing"
    tier        = "storage backbone"
  }
}

resource "azurerm_kubernetes_cluster_node_pool" "communications" {
  name                  = "np004"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.ragnarok.id
  vm_size               = var.np004_node_size
  enable_auto_scaling   = var.enable_auto_scaling_np004
  max_count             = var.max_node_count_np004
  min_count             = var.min_node_count_np004
  os_disk_size_gb       = var.np004_node_disk_size

  tags = {
    environment = "poc"
    project     = "trade matching system load testing"
    tier        = "communications backbone"
  }
}

resource "azurerm_kubernetes_cluster_node_pool" "producers" {
  name                  = "np005"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.ragnarok.id
  vm_size               = var.np005_node_size
  enable_auto_scaling   = var.enable_auto_scaling_np005
  max_count             = var.max_node_count_np005
  min_count             = var.min_node_count_np005
  os_disk_size_gb       = var.np005_node_disk_size

  tags = {
    environment = "poc"
    project     = "trade matching system load testing"
    tier        = "communications backbone"
  }
}


