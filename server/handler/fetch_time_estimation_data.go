package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kataras/tablewriter"
	"github.com/landoop/tableprinter"
	"github.com/IBM-Cloud/pulse-time-estimation-service/helpers"
	"github.com/IBM-Cloud/pulse-time-estimation-service/payload"
)

//GetTimeEstimationData ...
func GetTimeEstimationData(c *gin.Context) {

	fmt.Println("Inside GetTotalTimeEstimation")
	id := c.Param("id")

	jobID, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("ERROR in converting jobID: ", id)
	}

	if jobinfo, exist := payload.JobInfoMap[jobID]; exist {

		totalTime := convertToMinutes(jobinfo.TotalTimeEstimation)

		//Print Estimation details in table format
		printer := tableprinter.New(os.Stdout)

		// Optionally, customize the table, import of the underline 'tablewriter' package is required for that.
		printer.BorderTop, printer.BorderBottom, printer.BorderLeft, printer.BorderRight = true, true, true, true
		printer.CenterSeparator = "│"
		printer.ColumnSeparator = "│"
		printer.RowSeparator = "─"
		printer.HeaderBgColor = tablewriter.BgBlackColor
		printer.HeaderFgColor = tablewriter.FgGreenColor

		printer.Print(helpers.GetTable(jobinfo.ResourceList))
		fmt.Println("\nTotal Estimated Time: ", totalTime)

		resources := make([]payload.ResourceData, 0)

		for _, resource := range jobinfo.ResourceList {

			timeEstimation := convertToMinutes(resource.TimeEstimation)
			r := payload.ResourceData{
				ID:                 resource.ID,
				Name:               resource.Name,
				Region:             resource.Region,
				TimeEstimation:     timeEstimation,
				ServiceType:        resource.ServiceType,
				Action:             resource.Action,
				StartTime:          resource.StartTime,
				Day:                resource.Day,
				AccuracyPercentage: resource.AccuracyPercentage,
			}
			resources = append(resources, r)
		}

		s := strconv.Itoa(jobID)
		result := payload.TimeEstimationResult{
			ID:                  s,
			TotalTimeEstimation: totalTime,
			Resources:           resources,
		}

		byteArray, err := json.Marshal(result)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(string(byteArray))

		c.JSON(http.StatusOK, gin.H{
			"JobID":               id,
			"TotalTimeEstimation": totalTime,
			"Resources":           resources,
		})
	} else {
		c.JSON(http.StatusNotFound, map[string]string{
			"JobID":   id,
			"message": "JobId not found",
		})
	}
}

func convertToMinutes(totalSecs int64) string {
	hours := totalSecs / 3600
	minutes := (totalSecs % 3600) / 60
	seconds := totalSecs % 60

	timeString := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	return timeString
}
