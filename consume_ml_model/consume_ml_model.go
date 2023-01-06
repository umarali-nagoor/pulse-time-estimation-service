package ml

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var default_ml_url = "https://us-south.ml.cloud.ibm.com/ml/v4/deployments/c9d99b10-9b90-4638-829a-400ce9cf3510/predictions?version=2022-10-21"

// access_token for ml model
type AccessTokenResponse struct {
	Token string `json:"access_token"`
}

// data ...
type data struct {
	Fields []string    `json:"fields"`
	Values [][]float32 `json:"values"`
}

// response ...
type response struct {
	Predictions []data `json:"predictions"`
}

// final_estimation ...
type final_estimation struct {
	estimation_time float32
}

func generate_access_token() string {
	api_key := os.Getenv("IC_API_KEY")

	if api_key == "" {
		log.Fatal("Set IC_API_KEY")
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

	var access_token AccessTokenResponse
	json.Unmarshal(body_bytes, &access_token)
	fmt.Println("RETRIEVED ACCESS_TOKEN: ", access_token)

	return access_token.Token

}

// input the fields
//curl -X POST --header "Content-Type: application/json" --header "Accept: application/json" --header "Authorization: Bearer $IAM_TOKEN" -d '{"input_data": [{"fields": [$ARRAY_OF_INPUT_FIELDS],"values": [$ARRAY_OF_VALUES_TO_BE_SCORED, $ANOTHER_ARRAY_OF_VALUES_TO_BE_SCORED]}]}' "https://us-south.ml.cloud.ibm.com/ml/v4/deployments/4a5aed66-186d-44ee-976f-81fd54506b4a/predictions?version=2022-01-24"

func fetch_data_from_ml_model(access_token, resourceName, action, region, day, service, ml_api_endpoint string) {

	client := &http.Client{}
	log.Printf("Sending request :")

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

	var mlResp response
	err2 := json.Unmarshal([]byte(bodyText), &mlResp)
	if err2 != nil {
		panic(err2)
	}

	fmt.Printf("%+v\n", mlResp)
	fmt.Printf("%+v\n", mlResp.Predictions[0].Values[0][0])
}

// func main() {
// 	invoke_token := generate_access_token()
// 	log.Printf("invoke_token  %s", invoke_token)

// 	ml_api_endpoint := os.Getenv("ML_API_ENDPOINT")
// 	if ml_api_endpoint == "" {
// 		ml_api_endpoint = default_ml_url
// 		log.Printf("Defaulting ENDPOINT  to  %s", ml_api_endpoint)
// 	}
// 	fetch_data_from_ml_model(invoke_token, "ibm_resource_instance,", "create", "us_east", "Fri", "cloud-object-storage", ml_api_endpoint)
// }
