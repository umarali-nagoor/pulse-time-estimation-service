package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/IBM-Cloud/pulse-time-estimation-service/db"
	"github.com/IBM-Cloud/pulse-time-estimation-service/parser"
	"github.com/IBM-Cloud/pulse-time-estimation-service/payload"

	"github.com/gin-gonic/gin"
)

var defaultRegion string = "us-east"
var bufsize int = 5000000

// GetJobStatus ...
func GetJobStatus(context *gin.Context) {
	fmt.Println("Inside GetJobStatus")
	id := context.Param("id")

	jobID, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("ERROR in converting jobID: ", id)
	}

	if jobinfo, exist := payload.JobInfoMap[jobID]; exist {
		s := jobinfo.GetState()
		status := payload.ToString(s)
		context.JSON(http.StatusOK, map[string]interface{}{
			"jobID":  jobID,
			"Status": status,
		})
	} else {
		context.JSON(http.StatusOK, map[string]interface{}{
			"jobID":  jobID,
			"result": "JobID NotFound",
		})
	}

}

// PredictProvisionTime ...
func PredictProvisionTime(context *gin.Context) {
	// Read the Request Body content
	var bodyBytes []byte
	if context.Request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(context.Request.Body)
	}

	// Restore the io.ReadCloser to its original state
	context.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	buf := make([]byte, bufsize)
	num, _ := context.Request.Body.Read(buf)
	reqBody := string(buf[0:num])

	in := []byte(reqBody)
	var data map[string]interface{}
	if err := json.Unmarshal(in, &data); err != nil {
		panic(err)
	}

	//Crerate JobID randomly
	jobID := createJobID(1, 100)
	fmt.Println("Generated new JobId : ", jobID)

	job := new(payload.Job)
	job.ResourceList = make([]payload.ResourceInfo, 0)
	job.ResourceDependencyMap = make(map[string][]string, 0)
	job.TotalTimeEstimation = 0
	job.SetState(payload.INPRPGRESS)

	payload.JobInfoMap[jobID] = *job

	jobCreationResponse(context, jobID)

	fmt.Println("********* Received plan json : ", data)
	go processRequest(context, data, jobID)

}

func jobCreationResponse(context *gin.Context, jobID int) {

	context.JSON(http.StatusOK, map[string]interface{}{
		"jobID":  jobID,
		"result": "Job Created Successfully",
	})
}

func processRequest(context *gin.Context, data map[string]interface{}, jobID int) {
	fmt.Println("***** inside processRequest: ******", jobID)
	//time.Sleep(time.Second * 13)
	var jobinfo payload.Job
	var exist bool
	provisionTimePerResource := make(map[string]int64)

	var totalTimeEstimation int64

	resourceDependencyMap, startingNodes := parser.PrepareResourceDependecyList(data)

	//fmt.Printf("***** resourceDependencyList %v\n", resourceDependencyMap)
	//fmt.Printf("***** startingNodes %v\n", startingNodes)

	if jobinfo, exist = payload.JobInfoMap[jobID]; exist {
		jobinfo.ResourceDependencyMap = resourceDependencyMap
		jobinfo.StartingNodes = startingNodes
		payload.JobInfoMap[jobID] = jobinfo
	}

	fmt.Println("JObInfo after dependencylist set:", payload.JobInfoMap[jobID])

	/*
		resourceAttributeMap has format map[string]map[string]string

		ibm_resource_instance.cos_instance:         actions:create
												    location:global
												    service:cloud-object-storage

		ibm_cos_bucket.standard-ams03				actions:create
													location:global
	*/

	/*resourceInfo, err := db.GetAll()
	if err != nil {
		fmt.Println("ERROR in response for ALL resources:")
	}
	fmt.Printf("***** GteAll Resources %v\n", resourceInfo)*/

	resourceAttributeMap := parser.GetArgumentListPerResource(data)
	fmt.Printf("***** resourceAttributeMap %v\n", resourceAttributeMap)

	//Read region info from provider block if available
	var providerRegion = ""
	providerInfoMap := parser.GetProviderInfo(data)
	if configMap, ok := providerInfoMap["ibm"]; ok {
		if region, ok := configMap["region"]; ok {
			providerRegion = region.(string)
		}
	}

	// check for resource update
	var isUpdate = false
	updatedResources := make([]string, 0)
	if data["prior_state"] != nil {
		priorState := data["prior_state"].(map[string]interface{})
		if len(priorState) != 0 {
			isUpdate = true
			updatedResources = parser.GetUpdatedResourceList(data)
		}

		fmt.Printf("***** updatedResources %v\n", updatedResources)
	}
	fmt.Printf("***** isUpdate %v\n", isUpdate)

	var resourceList []payload.ResourceInfo

	//Sending request to DB
	for resource, value := range resourceAttributeMap {
		var action, location, service string

		//if its an update, request only for updated resources
		if isUpdate == true {
			var isPresent = false
			//resourceAttributeMap has all resoures and updatedResources has only updated resources
			//When there is an update we need to fetch data from db only for updated resources
			for _, r := range updatedResources {
				if r == resource {
					isPresent = true
					break
				}
			}

			if isPresent == false {
				continue
			}
		}

		//Get action
		if v, exist := value["actions"]; exist {
			action = v.(string)
		}

		//Set region from resource level, if not available set it from provider block if not available send default region "us-south"
		if v, exist := value["location"]; exist {
			location = v.(string)
		} else if providerRegion != "" {
			location = providerRegion
		} else {
			location = defaultRegion
		}

		//Get serviceType
		if v, exist := value["service"]; exist {
			service = v.(string)
		} else {
			service = "NA"
		}

		//Get Day
		day, err := getDay()
		if err != nil {
			return
		}

		//key is combination of ibm_iam_authorization_policy.policy
		resourceString, error := parser.SplitString(resource)

		if len(resourceString) == 0 || error != nil {
			fmt.Println("Error in spliting the resource:", resource)
			continue
		}

		fmt.Printf("########## Before sending resourceName: %s region: %s action: %s service: %s day: %s", resourceString[0], location, action, service, day)

		var resourceInfo *payload.ResourceInfo
		use_ml_model, provided := os.LookupEnv("USE_ML_MODEL")

		if !provided || (provided && use_ml_model == "true") {
			fmt.Println("Fetching data from ML Model")

			resourceInfo, err = db.GetEstimationUsingMLModel(resourceString[0], action, location, service, day)
			if err != nil {
				fmt.Println("ERROR in response for resource:", resource)
			}

			// if resourceInfo.ID == "" {
			// 	fmt.Println("NOT FOUND: Fetch from fallback DB ", resourceString[0])
			// 	// Right now, fallback db has data related to us-south, hence hard-coded
			// 	resourceInfo, err = db.GetOne(resourceString[0], action, defaultRegion, service, day, true)
			// 	if err != nil {
			// 		fmt.Println("NOT FOUND: in fallback DB:", resourceString[0])
			// 	}
			// }
		} else {
			fmt.Println("Fetching data from Fallback DB")
			// resourceInfo, err = db.GetOne(resourceString[0], action, defaultRegion, service, day, true)
			// if err != nil {
			// 	fmt.Println("NOT FOUND: in fallback DB:", resourceString[0])
			// }
		}

		provisionTimePerResource[resource] = resourceInfo.TimeEstimation

		resourceList = append(resourceList, *resourceInfo)

		/*fmt.Println("Response for resource:", resource[0])
		fmt.Printf("ID: %s\n", resourceInfo.ID)
		fmt.Printf("Name: %v\n", resourceInfo.Name)
		fmt.Printf("Region: %s\n", resourceInfo.Region)
		fmt.Printf("TimeEstimation: %v\n", resourceInfo.TimeEstimation)
		fmt.Printf("Action: %s\n", resourceInfo.Action)
		fmt.Printf("StartTime: %s\n", resourceInfo.StartTime)
		fmt.Printf("Day: %s\n", resourceInfo.Day)
		fmt.Printf("AccuracyPercentage: %v\n", resourceInfo.AccuracyPercentage)*/
	}

	if jobinfo, exist = payload.JobInfoMap[jobID]; exist {
		jobinfo.ResourceList = resourceList
		jobinfo.TimePerResourceMap = provisionTimePerResource
		payload.JobInfoMap[jobID] = jobinfo
		totalTimeEstimation = estimateTotalTime(jobID)
	}

	if jobinfo, exist = payload.JobInfoMap[jobID]; exist {
		fmt.Println("******** Setting Job Stat eto COMPLETED : ********** ")
		jobinfo.TotalTimeEstimation = totalTimeEstimation
		jobinfo.SetState(payload.COMPLETED)
		payload.JobInfoMap[jobID] = jobinfo
	} else {
		context.JSON(http.StatusOK, map[string]interface{}{
			"jobID":  jobID,
			"result": "JobID NotFound",
		})
	}

	fmt.Println("JObInfo after totalTimeEstimation set:", payload.JobInfoMap[jobID])

	fmt.Println("ProvisionTime: ", totalTimeEstimation)

}

// creates random id, Returns an int >= min, < max
func createJobID(min, max int) int {
	return min + rand.Intn(max-min)
}

func getDay() (string, error) {
	dt := time.Now()
	day := dt.Format("01-02-2006 15:04:05 Mon")
	if strings.Contains(day, " ") {
		tokens := strings.Split(day, " ")
		return tokens[2], nil
	}
	return " ", fmt.Errorf("The given idayd %s does not contain space", day)
}

/**********************************************************************************************************
1) startingNodes - list of resources on which no other resource depends on
 Say:

 cos_bucket
 service_policy
 cis_domain_settings

 2) resourceDependencyMap - For each resource we get the lit of resource on which it depends on

 cos_bucket            ------     resource_instance
							      kms_key

 service_policy        ------     service_id

 cis_domain_settings   ------     cis
								  cis_domain

3) provisionTimePerResource

  resource_instance    ------     5 sec
  kms_key              ------     2 sec
  service_id           ------     4 sec
  cis                  ------     7 sec
  cis_domain           ------     1 sec
  cos_bucket           ------     3 sec
  service_policy       ------     4 sec
  cis_domain_settings  ------     6 sec

  So to etimate say cos_bucket(3 sec) provision time we need to add resource_instance (5 sec) + kms_key(2 sec).
  Total estimation for cos_bucket is 3 + 5 + 2 = 10 sec

**********************************************************************************************************/

// func estimateTotalTime(startingNodes []string, resourceDependencyMap map[string][]string, provisionTimePerResource map[string]int64) int64 {
func estimateTotalTime(jobID int) int64 {
	startingNodes := payload.JobInfoMap[jobID].StartingNodes
	//resourceDependencyMap := payload.JobInfoMap[jobID].ResourceDependencyMap

	timeList := make([]int64, 0)

	for _, node := range startingNodes {
		timeList = append(timeList, totalTimeForStartingNode(node, jobID))
	}

	return max(timeList)
}

func totalTimeForStartingNode(resource string, jobID int) int64 {

	var resourceTime int64
	//startingNodes := payload.JobInfoMap[jobID].StartingNodes
	provisionTimePerResource := payload.JobInfoMap[jobID].TimePerResourceMap
	resourceDependencyMap := payload.JobInfoMap[jobID].ResourceDependencyMap

	var time int64
	var timeList []int64
	//first add starting node time in above e.g: cos_bucket time
	if t, present := provisionTimePerResource[resource]; present {
		time = time + t
	}
	//timeList = append(timeList, time)

	if depResources, found := resourceDependencyMap[resource]; found {
		//loop over all the dependent resources and sum them up
		for _, r := range depResources {
			resourceTime = totalTimeForStartingNode(r, jobID)
			timeList = append(timeList, resourceTime)
		}

	} else {
		return time
		//return max(timeList)
	}

	return time + max(timeList)
}

func max(resourceEstimationList []int64) int64 {
	var max int64
	if len(resourceEstimationList) > 0 {
		max = resourceEstimationList[0]
		for _, value := range resourceEstimationList {
			if value > max {
				max = value
			}
		}
	}
	return max
}

// Its a dummy logic for old DB model where each entry in db can have multiple resources
// func dummyLogicTotalTimeEstimation(jobId int, dependencyMap map[string][]string) int64 {
func dummyLogicTotalTimeEstimation(jobID int) int64 {
	var resourceEstimationList []int64
	if JobDetails, exist := payload.JobInfoMap[jobID]; exist {
		fmt.Printf("JObID: %v resourceList: %v", jobID, JobDetails.ResourceList)
		for resource, dependentList := range JobDetails.ResourceDependencyMap {
			fmt.Printf("resource %v dependentList %v", resource, dependentList)
			var totalTime int64
			//resource name vs list of resource names on which it dependent
			for _, r1 := range dependentList {
				for _, r2 := range JobDetails.ResourceList {
					r, _ := parser.SplitString(r1)
					if r[0] == r2.Name {
						totalTime = totalTime + r2.TimeEstimation
					}
				}
			}

			res, _ := parser.SplitString(resource)
			for _, r := range JobDetails.ResourceList {
				if r.Name == res[0] {
					totalTime = totalTime + r.TimeEstimation
				}
			}

			resourceEstimationList = append(resourceEstimationList, totalTime)

		}
		fmt.Printf("resourceEstimationList %v \n", resourceEstimationList)

		var max int64
		if len(resourceEstimationList) > 0 {
			max = resourceEstimationList[0]
			for _, value := range resourceEstimationList {

				if value > max {
					max = value
				}
			}
		}

		return max

	} else {
		fmt.Println("JobId does not exists", jobID)
	}

	return 0
}

/*
func dummyLogicTotalTimeEstimation(jobId int) int64 {
	var resourceEstimationList []int64
	if resourceList, exist := payload.JobInfoMap[jobId]; exist {
		fmt.Println("resourceList", resourceList)
		for resource, dependentList := range dependencyMap {
			fmt.Printf("resource %v dependentList %v", resource, dependentList)
			var totalTime int64
			//resource name vs list of resource names on which it dependent
			for _, r1 := range dependentList {
				for _, r2 := range resourceList {
					r, _ := parser.SplitString(r1)
					if len(r2.ResourceList) > 1 && r[0] == r2.ResourceList[1] {
						fmt.Printf("case 1 adding %v \n", r2.ElapsedTime)
						totalTime = totalTime + r2.ElapsedTime
					} else if len(r2.ResourceList) == 1 && r2.ResourceList[0] == r[0] {
						fmt.Printf("case 2 adding %v \n", r2.ElapsedTime)
						totalTime = totalTime + r2.ElapsedTime
					}
				}
			}

			res, _ := parser.SplitString(resource)
			for _, r := range resourceList {
				if r.ResourceList[0] == res[0] {
					fmt.Printf("case 3 adding %v \n", r.ElapsedTime)
					totalTime = totalTime + r.ElapsedTime
				}
			}

			resourceEstimationList = append(resourceEstimationList, totalTime)

		}
		fmt.Printf("resourceEstimationList %v \n", resourceEstimationList)

		max := resourceEstimationList[0]
		for _, value := range resourceEstimationList {

			if value > max {
				max = value
			}
		}

		return max

	} else {
		fmt.Println("JobId does not exists", jobId)
	}

	return 0
}
*/
