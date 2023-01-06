package resource

type ResourceInfo struct {
	ResourceName string `json:"resource_name"`
	ResourceID   string `json:"service_id"`
	Region       string `json:"region"`
	Day          string `json:"day"`
	Time         string `json:"time"`
}

type AllResources struct {
	Resources []ResourceInfo
}

type Response struct {
	EstimationTime      float64 `json:"time"`
	ResourceID          string  `json:"resource_id"`
	TotalEstimationTime float64 `json:"total_time"`
}

//Used create new resource list
func New() *AllResources {
	return &AllResources{
		Resources: []ResourceInfo{},
	}
}

func (r *AllResources) AddResource(newResource ResourceInfo) {
	r.Resources = append(r.Resources, newResource)
}

//returns all resources
func (r *AllResources) GetAllResources() []ResourceInfo {
	return r.Resources
}
