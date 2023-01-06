resource "ibm_is_vpc" "testacc_vpc" {
  name = "myvpc-2"
}

resource "ibm_is_security_group" "testacc_security_group" {
  name = "mysg-2"
  vpc  = ibm_is_vpc.testacc_vpc.id
}
