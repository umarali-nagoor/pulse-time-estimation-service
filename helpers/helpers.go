package helpers

import (
	"fmt"

	"github.com/IBM-Cloud/pulse-time-estimation-service/payload"
)

//DisplayTable ...
type DisplayTable struct {
	Resource           string `header:"name" bson:"name"`
	Region             string `header:"region" bson:"region"`
	TimeEstimation     string `header:"time_Estimation"`
	ServiceType        string `header:"service_Type"`
	Action             string `header:"action"`
	Day                string `header:"day"`
	AccuracyPercentage int64  `header:"confidence_score"`
}

//GetTable ...
func GetTable(resources []payload.ResourceInfo) []DisplayTable {
	resMap := make([]DisplayTable, 0, len(resources))

	for _, resource := range resources {
		var instance DisplayTable
		instance.Resource = resource.Name
		instance.Region = resource.Region
		instance.TimeEstimation = ConvertToMinutes(resource.TimeEstimation)
		instance.ServiceType = resource.ServiceType
		instance.Action = resource.Action
		instance.Day = resource.Day
		instance.AccuracyPercentage = resource.AccuracyPercentage
		resMap = append(resMap, instance)

	}
	return resMap
}

//ConvertToMinutes ...
func ConvertToMinutes(totalSecs int64) string {
	hours := totalSecs / 3600
	minutes := (totalSecs % 3600) / 60
	seconds := totalSecs % 60

	timeString := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	return timeString
}
