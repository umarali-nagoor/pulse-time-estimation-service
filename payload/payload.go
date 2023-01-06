package payload

import "fmt"

type resources []ResourceInfo

// JobInfo - captures all resources info against a job
//var JobInfo = make(map[int][]ResourceInfo)

// JobInfo - captures all resources info against a job
var JobInfoMap = make(map[int]Job)

//ResourceInfo  Used to capture data from DB.
type ResourceInfo struct {
	ID                 string `json:"id,omitempty" bson:"_id,omitempty"`
	Name               string `json:"name" bson:"name"`
	Region             string `json:"region" bson:"region"`
	TimeEstimation     int64  `bson:"timeEstimation"`
	ServiceType        string `bson:"serviceType"`
	Action             string `bson:"action"`
	StartTime          string `bson:"startTime"`
	Day                string `bson:"day"`
	AccuracyPercentage int64  `bson:"accuracyPercentage"`
}

//ResourceData  Used to send time estimation data results.
type ResourceData struct {
	ID                 string `json:"id,omitempty" bson:"_id,omitempty"`
	Name               string `json:"name" bson:"name"`
	Region             string `json:"region" bson:"region"`
	TimeEstimation     string `bson:"timeEstimation"`
	ServiceType        string `bson:"serviceType"`
	Action             string `bson:"action"`
	StartTime          string `bson:"startTime"`
	Day                string `bson:"day"`
	AccuracyPercentage int64  `bson:"accuracyPercentage"`
}

// TimeEstimationResult used to output the result
type TimeEstimationResult struct {
	ID                  string         `json:"id" bson:"id"`
	TotalTimeEstimation string         `json:"totalTimeEstimation" bson:"totalTimeEstimation"`
	Resources           []ResourceData `json:"resources" bson:"resources"`
}

// Job has the complete resource info and total time estimation
type Job struct {
	ResourceList          []ResourceInfo
	ResourceDependencyMap map[string][]string
	StartingNodes         []string
	TimePerResourceMap    map[string]int64
	TotalTimeEstimation   int64
	State                 JobState
}

//SetState ...
func (job *Job) SetState(state JobState) {
	job.State = state
}

//GetState ...
func (job Job) GetState() JobState {
	return job.State
}

// JobState ...
type JobState int

//ToString ... convert int enum to string
func ToString(s JobState) string {
	switch s {
	case INITIAL:
		return "INITIAL"
	case INPRPGRESS:
		return "INPRPGRESS"
	case COMPLETED:
		return "COMPLETED"
	case FAILED:
		return "FAILED"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}

// states ...
const (
	INITIAL JobState = 1 + iota
	INPRPGRESS
	COMPLETED
	FAILED
)

//AccessTokenResponse ...
type AccessTokenResponse struct {
	Token string `json:"access_token"`
}

//Data ...
type Data struct {
	Fields []string    `json:"fields"`
	Values [][]float32 `json:"values"`
}

//response ...
type Response struct {
	Predictions []Data `json:"predictions"`
}

//Final_estimation ...
type Final_estimation struct {
	estimation_time float32
}

// {
// 	"predictions": [{
// 	  "fields": ["prediction"],
// 	  "values": [[18.602850579090877]]
// 	}]
// }

// {
// 	"input_data": [
// 			{
// 					"fields": [
// 							"action",
// 							"day",
// 							"name",
// 							"region",
// 							"service_type"
// 					],
// 					"values": []
// 			}
// 	]
// }

//InnerData ...
// type InnerData struct {
// 	Fields []string   `json:"fields"`
// 	Values [][]string `json:"values"`
// }

// //OuterData ...
// type OuterData struct {
// 	Input_data []InnerData `json:"input_data"`
// }
