provider "ibm" {
  generation = 1
}

#*********************************************
# Resource group creation
#*********************************************
resource "ibm_resource_group" "demo_group1" {
  name     = "prod"
}

#*********************************************
#KMS
#*********************************************
resource "ibm_resource_instance" "kp_instance" {
  name     = "demo_KMS_instance"
  service  = "kms"
  plan     = "tiered-pricing" 
  location = "us-south"
}

resource "ibm_kp_key" "cos_encrypt" {
  key_protect_id  = ibm_resource_instance.kp_instance.guid
  key_name     = "key-name"
  standard_key = false
}

#****************************************
#provision CIS
#*****************************************

resource "ibm_cis" "demo_web_domain" {
  name              = "web_domain"
  resource_group_id = ibm_resource_group.demo_group1.id
  plan              = "standard"
  location          = "global"
}

resource "ibm_cis_domain_settings" "demo_web_domain" {
  cis_id          = ibm_cis.demo_web_domain.id
  domain_id       = ibm_cis_domain.demo_web_domain.id
  waf             = "on" #set this off to trigger an alert
  ssl             = "full"
  min_tls_version = "1.2"
}

resource "ibm_cis_domain" "demo_web_domain" {
  cis_id = ibm_cis.demo_web_domain.id
  domain = "demo.ibm.com"
}

#*****************************************
#provision COS
#*****************************************

resource "ibm_resource_instance" "cos_instance" {
  name              = "demo_cos_instance"
  resource_group_id = ibm_resource_group.demo_group1.id
  service           = "cloud-object-storage"
  plan              = "standard"
  location          = "global"
}

resource "ibm_cos_bucket" "standard-ams03" {
  bucket_name          = "terraform-demo-bucket-m98hji6hgk89067ga"
  resource_instance_id = ibm_resource_instance.cos_instance.id
  region_location      = "us-south"
  storage_class        = "standard"
  key_protect          = ibm_kp_key.cos_encrypt.id
}

#******************************************
# setup IAM
#*****************************************

#auth policy for cos to read kms keys
resource "ibm_iam_authorization_policy" "policy" {
  source_service_name         = "cloud-object-storage"
  target_service_name         = "kms"
  roles                       = ["Reader"]
}

#*****************************************
#user policies
#*****************************************

resource "ibm_iam_user_policy" "policy1" {
  ibm_id = var.user1
  roles  = ["Viewer", "Administrator"]

  resources  {
    service = "kms"
  }
 
}

resource "ibm_iam_user_policy" "policy2" {
  ibm_id = var.user2
  roles  = ["Viewer"]

  resources  {
    service = "kms"
  }
}

#*****************************************
#service id
#*****************************************
resource "ibm_iam_service_id" "serviceID" {
  name = "demo-cis-dervice"
}

resource "ibm_iam_service_policy" "policy" {
  iam_service_id = "${ibm_iam_service_id.serviceID.id}"
  roles        = ["Writer"]

  resources { 
    service = "cloud-object-storage"
    resource_group_id = "demo_group1" 
  }
}

#**********************************************
#VPC Cluster
#**********************************************

resource "ibm_is_vpc" "vpc1" {
  name = "myvpc"
}

resource "ibm_is_subnet" "subnet1" {
  name                     = "mysubnet1"
  vpc                      = ibm_is_vpc.vpc1.id
  zone                     = "us_south-1"
  total_ipv4_address_count = 256
}

resource "ibm_is_subnet" "subnet2" {
  name                     = "mysubnet2"
  vpc                      = ibm_is_vpc.vpc1.id
  zone                     = "us-south-2"
  total_ipv4_address_count = 256
}

resource "ibm_container_vpc_cluster" "cluster" {
  name              = "mycluster"
  vpc_id            = ibm_is_vpc.vpc1.id
  flavor            = "bc1-2x8"
  worker_count      = 3
  resource_group_id = ibm_resource_group.demo_group1.id

  zones {
    subnet_id = ibm_is_subnet.subnet1.id
    name      = "us-south-1"
  }
}

resource "ibm_container_bind_service" "bind_service" {
  cluster_name_id     = ibm_container_vpc_cluster.cluster.id
  service_instance_id = element(split(":", ibm_resource_instance.cos_instance.id), 7)
  namespace_id        = "default"
  role                = "Writer"
}
