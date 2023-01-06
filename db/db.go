package db

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/IBM-Cloud/pulse-time-estimation-service/payload"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var db *mgo.Database
var fallbckdb *mgo.Database

// DbName ...
var DbName = "fallback_db"
var collName = "fallback_collection"
var certPath = "../ca-certificate.crt"

// FallbackDBName ...
var FallbackDBName = "fallback_db"
var fallbackCollName = "fallback_collection"

var default_ml_url = "https://us-south.ml.cloud.ibm.com/ml/v4/deployments/c9d99b10-9b90-4638-829a-400ce9cf3510/predictions?version=2022-10-21"

// ConnectToDB is to connect to db
func ConnectToDB(dbName string, fallback bool) {

	rootPEM, err := ioutil.ReadFile(certPath)
	roots := x509.NewCertPool()
	_ = roots.AppendCertsFromPEM([]byte(rootPEM))

	tlsConfig := &tls.Config{
		RootCAs:            roots,
		InsecureSkipVerify: true}

	db_url := os.Getenv("DB_URL")
	if db_url == "" {
		log.Fatal("ERROR: DB_URL is not set")
	}
	log.Printf("DB_URL  %s", db_url)

	dialInfo, err := mgo.ParseURL(db_url)
	if err != nil {
		log.Println(err)
	}

	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
		if err != nil {
			log.Println(err)
		}
		return conn, err
	}

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		panic(err)
	}

	//defer session.Close()
	if fallback == false {
		db = session.DB(dbName)
	} else {
		fallbckdb = session.DB(dbName)
	}

}

func getCollection(fallback bool) *mgo.Collection {

	/*Try to read environment set values, if they are empty then read deafult values*/

	if fallback == true {

		fallbackCollection := os.Getenv("FALLBACK_COLLECTION")

		if fallbackCollection == "" {
			fallbackCollection = fallbackCollName
			log.Printf("Defaulting fallbackCollection to  %s", fallbackCollection)
		}
		return fallbckdb.C(fallbackCollection)

	} else {

		primaryCollection := os.Getenv("PRIMARY_COLLECTION")

		if primaryCollection == "" {
			primaryCollection = collName
			log.Printf("Defaulting primaryCollection to  %s", primaryCollection)
		}
		return db.C(primaryCollection)

	}
}

// GetAll returns all items from the requested database.
func GetAll() ([]payload.ResourceInfo, error) {

	res := []payload.ResourceInfo{}

	//send false to indicate that its not for fallback db
	if err := getCollection(false).Find(nil).All(&res); err != nil {
		return nil, err
	}

	return res, nil
}

// GetOne returns a single item from the database.
func GetOne(resourceName, action, region, service, day string, fallback bool) (*payload.ResourceInfo, error) {
	res := payload.ResourceInfo{}

	fmt.Printf("resourceName: %s region: %s action: %s service: %s day: %s \n", resourceName, region, action, service, day)

	if err := getCollection(fallback).Find(bson.M{
		"$and": []bson.M{
			{"name": resourceName},
			{"region": region},
			{"action": action},
			{"day": day},
			{"serviceType": service},
		},
	}).One(&res); err != nil {
		fmt.Println("  ERROR in GetOne  ")
	}

	return &res, nil
}

// GetEstimationUsingMLModel ...
func GetEstimationUsingMLModel(resourceName, action, region, service, day string) (*payload.ResourceInfo, error) {

	Token := generate_access_token()

	ml_api_endpoint := os.Getenv("ML_API_ENDPOINT")
	if ml_api_endpoint == "" {
		log.Println("ERROR: ML_API_ENDPOINT is not set, using default value")
		ml_api_endpoint = default_ml_url
	}
	log.Printf("ML_API_ENDPOINT  %s", ml_api_endpoint)

	time_estimation := fetch_data_from_ml_model(Token, resourceName, action, region, service, day, ml_api_endpoint)
	fmt.Println("Time Estimation from ML Model: ", time_estimation)

	res := payload.ResourceInfo{
		ID:                 "NA",
		Name:               resourceName,
		Region:             region,
		TimeEstimation:     int64(time_estimation),
		ServiceType:        service,
		Action:             action,
		StartTime:          "NA",
		Day:                day,
		AccuracyPercentage: 70,
	}

	return &res, nil

}

// generate_access_token ...
func generate_access_token() string {
	api_key := os.Getenv("IC_API_KEY")

	if api_key == "" {
		log.Fatal("ERROR: IC_API_KEY is not set")
		return ""
	}

	client := &http.Client{}
	var data = strings.NewReader(`grant_type=urn:ibm:params:oauth:grant-type:apikey&apikey=` + api_key)

	req, err := http.NewRequest("POST", "https://iam.cloud.ibm.com/identity/token", data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body_bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", body_bytes)

	var access_token payload.AccessTokenResponse
	json.Unmarshal(body_bytes, &access_token)
	//fmt.Println("RETRIEVED ACCESS_TOKEN: ", access_token)

	return access_token.Token

}

// input the fields
//curl -X POST --header "Content-Type: application/json" --header "Accept: application/json" --header "Authorization: Bearer $IAM_TOKEN" -d '{"input_data": [{"fields": [$ARRAY_OF_INPUT_FIELDS],"values": [$ARRAY_OF_VALUES_TO_BE_SCORED, $ANOTHER_ARRAY_OF_VALUES_TO_BE_SCORED]}]}' "https://us-south.ml.cloud.ibm.com/ml/v4/deployments/4a5aed66-186d-44ee-976f-81fd54506b4a/predictions?version=2022-01-24"

// fetch_data_from_ml_model ...
func fetch_data_from_ml_model(access_token, resourceName, action, region, service, day, ml_api_endpoint string) float32 {

	client := &http.Client{}
	log.Printf("Sending request :")

	//var body = strings.NewReader(`{"input_data": [{"fields": ["action","day","name","region","service_type"],"values": [[action,day,resourceName,region,service]]}]}`)
	var body = strings.NewReader(`{"input_data": [{"fields": ["action","day","name","region","service_type"],"values": [["create","Fri","ibm_resource_instance","us-east","cloud-object-storage"]]}]}`)
	req, err := http.NewRequest("POST", ml_api_endpoint, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+access_token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	//output looks like
	// {
	// 	"predictions": [{
	// 	  "fields": ["prediction"],
	// 	  "values": [[18.602850579090877]]
	// 	}]
	// }

	fmt.Printf("%s\n", bodyText)

	var mlResp payload.Response
	err2 := json.Unmarshal([]byte(bodyText), &mlResp)
	if err2 != nil {
		panic(err2)
	}

	time_estimation := mlResp.Predictions[0].Values[0][0]

	fmt.Printf("%+v\n", mlResp)
	fmt.Printf("%+v\n", time_estimation)

	return time_estimation
}
