package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var db *mgo.Database

//resource represents a sample database entity.
type resource struct {
	ID           string   `json:"id,omitempty" bson:"_id,omitempty"`
	ResourceList []string `json:"resources" bson:"resource"`
	Region       string   `json:"region" bson:"region"`
	ElapsedTime  int64    `bson:"ElapsedTime"`
	actions      string   `bson:"type"`
	BeginTime    string   `bson:"begin_timestamp"`
}

func main() {
	rootPEM, err := ioutil.ReadFile("/Users/umarali/go/src/github.com/IBM-Cloud/resource-provision-estimation-time/ca-certificate.crt")
	roots := x509.NewCertPool()
	_ = roots.AppendCertsFromPEM([]byte(rootPEM))

	tlsConfig := &tls.Config{
		RootCAs:            roots,
		InsecureSkipVerify: true}
	dialInfo, err := mgo.ParseURL("mongodb://admin:Kavya12345@cc420766-a319-4c3e-b183-38a6019b4c4f-0.bmo1leol0d54tib7un7g.databases.appdomain.cloud:32760,cc420766-a319-4c3e-b183-38a6019b4c4f-1.bmo1leol0d54tib7un7g.databases.appdomain.cloud:32760/ibmclouddb?authSource=admin&replicaSet=replset")
	if err != nil {
		log.Println(err)
	}
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		//tlsConfig := &tls.Config{}
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
	defer session.Close()
	fmt.Printf("Connected to %v!\n", session.LiveServers())
	dbNames, err := session.DatabaseNames()
	if err != nil {
		fmt.Println("DB List Error")
	}
	fmt.Printf("DB List %v!\n", dbNames)
	fmt.Printf("DB %v!\n", session.DB("action_data"))

	db = session.DB("action_data")
	cNames, err := db.CollectionNames()
	fmt.Printf("%%%%%%% Collection List %v!\n", cNames)

	curatedCollection := getCollection("curated_data")
	estimatedCollection := getCollection("estimated_time")

	//Get Collecgion in one line,  coll := session.DB("action_data").C("estimated_time")
	fmt.Printf("curatedCollection %v!\n", curatedCollection)
	fmt.Printf("estimatedCollection %v!\n", estimatedCollection)

	resources, error := GetAll("estimated_time")
	if error != nil {
		fmt.Printf("Error in GetALL %v!\n", error)
	}
	fmt.Printf("******* Total resources %v ************* \n", len(resources))
	for _, v := range resources {
		fmt.Printf("ID: %s\n", v.ID)
		fmt.Printf("Resource: %v\n", v.ResourceList)
		fmt.Printf("Region: %s\n", v.Region)
		fmt.Printf("ElapsedTime: %v\n", v.ElapsedTime)
		fmt.Printf("actions: %s\n", v.actions)
		fmt.Printf("BeginTime: %s\n", v.BeginTime)
		fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++")
	}

	fmt.Println("******* Resource Info ************* ")

	v, error := GetOne("estimated_time")
	if error != nil {
		fmt.Printf("Error in GetOne %v!\n", error)
	}
	fmt.Printf("ID: %s\n", v.ID)
	fmt.Printf("Resource: %v\n", v.ResourceList)
	fmt.Printf("Region: %s\n", v.Region)
	fmt.Printf("ElapsedTime: %v\n", v.ElapsedTime)
	fmt.Printf("actions: %s\n", v.actions)
	fmt.Printf("BeginTime: %s\n", v.BeginTime)

}

func getCollection(collectionName string) *mgo.Collection {
	return db.C(collectionName)
}

// GetAll returns all items from the database.
func GetAll(coll string) ([]resource, error) {

	res := []resource{}

	if err := getCollection(coll).Find(nil).All(&res); err != nil {
		return nil, err
	}

	return res, nil
}

// GetOne returns a single item from the database.
func GetOne(coll string) (*resource, error) {
	res := resource{}

	if err := getCollection(coll).Find(bson.M{
		"$and": []bson.M{
			{"resource": "ibm_iam_service_policy"},
			{"resource": "ibm_iam_service_id"},
			{"region": "us-south"},
			{"type": "create"},
		},
	}).One(&res); err != nil {
		fmt.Println("******* ERROR in GetOne ************* ")
	}

	/*if err := getCollection("estimated_time").Find(bson.M{"resource": [ibm_api_gateway_endpoint ibm_resource_instance]}).One(&res); err != nil {
		return nil, err
	}*/

	return &res, nil
}
