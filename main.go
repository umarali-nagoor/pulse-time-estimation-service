package main

import (
	"log"
	"os"

	"github.com/IBM-Cloud/pulse-time-estimation-service/server/handler"

	"github.com/IBM-Cloud/pulse-time-estimation-service/db"

	"github.com/gin-gonic/gin"
)

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	router := gin.Default()
	api := router.Group("/api")
	v1 := api.Group("/v1")

	v1.POST("/predictor", handler.PredictProvisionTime)

	v1.GET("/predictor/:id", handler.GetTimeEstimationData)

	v1.DELETE("/predictor/:id", handler.DeleteJob)

	v1.GET("/predictor/:id/status", handler.GetJobStatus)

	primaryDB := os.Getenv("PRIMARY_DB")

	if primaryDB == "" {
		primaryDB = db.DbName

	}

	fallbackDB := os.Getenv("FALLBACK_DB")

	if fallbackDB == "" {
		fallbackDB = db.FallbackDBName
	}

	log.Printf("Reading from PrimaryDB %s & FallbackDB  %s", primaryDB, fallbackDB)

	//connect to PrimaryDB
	//db.ConnectToDB(primaryDB, false)

	//connect to fallbackDB
	//db.ConnectToDB(fallbackDB, true)

	// listen and serve on 0.0.0.0:8080
	log.Printf("Listening on port %s", port)
	router.Run(":" + port)

}
