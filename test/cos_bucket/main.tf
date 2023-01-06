data "ibm_resource_group" "cos_group" {
  is_default = true
}

resource "ibm_resource_instance" "instance2" {
  name              = "cos1"
  resource_group_id = data.ibm_resource_group.cos_group.id
  service           = "cloud-object-storage"
  plan              = "standard"
  location          = "global"
}



# resource "ibm_resource_instance" "activity_tracker2" {
#   name              = "at"
#   resource_group_id = data.ibm_resource_group.cos_group.id
#   service           = "logdnaat"
#   plan              = "7-day"
#   location          = "us-south"
# }
# resource "ibm_resource_instance" "metrics_monitor2" {
#   name              = "mon"
#   resource_group_id = data.ibm_resource_group.cos_group.id
#   service           = "sysdig-monitor"
#   plan              = "graduated-tier"
#   location          = "us-south"
#   parameters = {
#     default_receiver = true
#   }
# }
# resource "ibm_cos_bucket" "bucket2" {
#   bucket_name          = "bucket"
#   resource_instance_id = ibm_resource_instance.instance2.id
#   single_site_location = "ams03"
#   storage_class        = "standard"
#   activity_tracking {
#     read_data_events     = true
#     write_data_events    = true
#     activity_tracker_crn = ibm_resource_instance.activity_tracker2.id
#   }
#   metrics_monitoring {
#     usage_metrics_enabled   = true
#     request_metrics_enabled = true
#     metrics_monitoring_crn  = ibm_resource_instance.metrics_monitor2.id
#   }
# }
