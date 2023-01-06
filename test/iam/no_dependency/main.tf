resource "ibm_iam_authorization_policy" "policy" {
  source_service_name = "cloud-object-storage"
  target_service_name = "kms"
  roles               = ["Reader"]
}

resource "ibm_iam_service_id" "serviceID" {
  name = "test"
}


