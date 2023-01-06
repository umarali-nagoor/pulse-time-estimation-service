resource "ibm_iam_service_id" "serviceID" {
  name = "test"
}


resource "ibm_iam_service_policy" "policy" {
  iam_service_id = ibm_iam_service_id.serviceID.id
  roles          = ["Reader","Manager"]

  resources {
    service = "cloud-object-storage"
  }
}

