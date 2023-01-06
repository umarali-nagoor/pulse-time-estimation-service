locals {
  bind = true
}
provider "ibm" {
}

data "ibm_resource_group" "cos_group" {
  name = "Default"
}

resource "ibm_resource_instance" "cos_instance" {
  count             = 1
  name              = "Test-cos"
  resource_group_id = data.ibm_resource_group.cos_group.id
  service           = "cloud-object-storage"
  plan              = "standard"
  location          = "global"
}

resource "ibm_resource_key" "key" {
  count                = local.bind ? 1 : 0
  name                 = "rk"
  role                 = "Manager"
  parameters           = {
    "HMAC" = true
  }
  resource_instance_id = ibm_resource_instance.cos_instance[0].id
}
