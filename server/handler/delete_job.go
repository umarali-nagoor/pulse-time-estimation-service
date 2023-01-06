package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/IBM-Cloud/pulse-time-estimation-service/payload"
)

//DeleteJob deletes a job
func DeleteJob(c *gin.Context) {
	fmt.Println("Inside DeleteJob")
	id := c.Param("id")

	jobID, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("ERROR in converting jobID: ", id)
	}

	delete(payload.JobInfoMap, jobID)
	_, present := payload.JobInfoMap[jobID]
	if present == false {
		c.JSON(http.StatusOK, map[string]string{
			"jobId":  id,
			"result": "Deleted Successfully",
		})
	}
}
