resource "ibm_iam_authorization_policy" "policy" {
  source_service_name = "cloud-object-storage"
  target_service_name = "kms"
  roles               = ["Reader"]
}

resource "ibm_iam_service_id" "serviceID" {
  name = "test"
}

resource "ibm_cos_bucket" "standard-ams03" {
  bucket_name          = "testbucket"
  resource_instance_id = "123"
  cross_region_location      = "us"
  storage_class        = "standard"
}
