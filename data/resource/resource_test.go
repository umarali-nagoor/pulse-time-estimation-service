package resource

import "testing"

func TestAddResource(t *testing.T) {
	resourceList := New()
	resourceList.AddResource(ResourceInfo{ResourceName: "R1", ResourceID: "1", Region: "us-south", Day: "Monday", Time: "12"})
	resourceList.AddResource(ResourceInfo{ResourceName: "R2", ResourceID: "2", Region: "us-south", Day: "Tuesday", Time: "15"})
	if len(resourceList.Resources) != 2 {
		t.Errorf("Resource was not added")
	}
}

func GetAllResources(t *testing.T) {
	resourceList := New()
	resourceList.AddResource(ResourceInfo{ResourceName: "R1", ResourceID: "1", Region: "us-south", Day: "Monday", Time: "12"})
	resourceList.AddResource(ResourceInfo{ResourceName: "R2", ResourceID: "2", Region: "us-south", Day: "Tuesday", Time: "15"})
	r := resourceList.GetAllResources()
	if len(r) != 2 {
		t.Errorf("Resource list was not updated")
	}
}
