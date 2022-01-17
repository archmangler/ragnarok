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
    name            = "np001"
    node_count      = var.np001_node_count
    vm_size         = var.np001_node_size //"Standard_D2_v2"
    os_disk_size_gb = var.np001_node_disk_size
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
