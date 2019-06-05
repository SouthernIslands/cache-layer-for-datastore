package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const baseURL string = "https://datastore.googleapis.com/v1/projects/"
const currentToken string = "ya29.GlwVBxUf84OM4JRMb1t8CNt92LoNZHyuqTdMj6lb3oEjxPLWbVrdgocOl6mXYObDfLZT6t5bL-SAQ5A9RHmbJNFNFtI_9guiL7t-Dpm-MruqGKe124WLm3jS6Fllog"
const projectid = "central-binder-241522"

type entry struct {
	key    string
	value  []byte
	expire time.Time
}

type Book struct {
	ID            int64
	Title         string
	Author        string
	PublishedDate string
	ImageURL      string
	Description   string
	CreatedBy     string
	CreatedByID   string
}

func main() {
	//DoGet()
	DoSet()
}

func DoGet() {
	message := map[string]interface{}{
		"keys": map[string]interface{}{
			"path": map[string]interface{}{
				"kind": "Book",
				"id":   5634472569470976,
			},
		},
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(baseURL+projectid+":lookup"+"?access_token="+currentToken,
		"application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	log.Println(resp)
	log.Println(result)
	if resp.StatusCode == http.StatusOK {
		if result["found"] != nil {
			//write int map
			//t := &entry{
			//	key:"kind"+"id",
			//	value:bytesRepresentation,
			//	expire:time.Now(),
			//}

			log.Println("Entity Found: ", result["found"])
		}
		if result["missing"] != nil {
			log.Println("Entity Not Found: ", result["missing"])
		}
		//res := http.ResponseWriter(resp)
		//past the response to client
		//outdated？
	} else {
		log.Println("WARNING: ", resp.StatusCode, resp.Body)
	}
}

func DoSet() {
	//begintrans
	transRequest := map[string]interface{}{
		"transactionOptions": map[string]interface{}{
			"readWrite": map[string]interface{}{},
		},
	}

	bytesRepresentation, err := json.Marshal(transRequest)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(baseURL+projectid+":beginTransaction"+"?access_token="+currentToken,
		"application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	var beginTransRespon map[string]string
	json.NewDecoder(resp.Body).Decode(&beginTransRespon)
	log.Println("Response: ", resp)
	log.Println(beginTransRespon)

	var trans string
	if resp.StatusCode == http.StatusOK {
		trans = beginTransRespon["transaction"]
	}

	tmp := &Book{
		ID:            1234567891,
		Title:         "Deng Xiaoping: The Man who Made Modern China",
		Author:        "Michael Dillon",
		PublishedDate: "2014",
		//ImageURL:      imageURL,
		//Description:   r.FormValue("description"),
		CreatedBy:   "Gopher",
		CreatedByID: "0",
	}

	//commit
	commitMessage := map[string]interface{}{
		"transaction": trans,
		"mutations": map[string]map[string]interface{}{
			"upsert": {
				"key": map[string]interface{}{
					"path": map[string]interface{}{
						"id":   56294995300001,
						"kind": "Book",
					},
				},
				"properties": map[string]interface{}{
					"PulishedDate": map[string]interface{}{
						"stringValue": tmp.PublishedDate,
					},
					"ImageURL": map[string]interface{}{
						"stringValue": "",
					},
					"Description": map[string]interface{}{
						"stringValue": "",
					},
					"CreatedBy": map[string]interface{}{
						"stringValue": tmp.CreatedBy,
					},
					"ID": map[string]interface{}{
						"integerValue": tmp.ID,
					},
					"Author": map[string]interface{}{
						"stringValue": tmp.Author,
					},
					"CreatedByID": map[string]interface{}{
						"stringValue": tmp.CreatedByID,
					},
					"Title": map[string]interface{}{
						"stringValue": tmp.Title,
					},
				},
			},
		},
	}

	bytesRepresentation, err = json.Marshal(commitMessage)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err = http.Post(baseURL+projectid+":commit"+"?access_token="+currentToken,
		"application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	log.Println("Response: ", resp)
	log.Println(result)

	if resp.StatusCode != http.StatusOK {
		//do nothing
		//reply to client
		return
	}

	//[]interface{} -> interface{} -> map
	mutations := result["mutationResults"].([]interface{})
	var t interface{}
	for _, i := range mutations {
		t = i
	}
	t2 := t.(map[string]interface{})
	version := t2["version"]

	//form a found result into map
	muta := commitMessage["mutations"].(map[string]map[string]interface{})
	key := muta["upsert"]["key"].(map[string]interface{})
	properties := muta["upsert"]["properties"].(map[string]interface{})
	key["partitionId"] = map[string]string{"projectId": projectid}

	entity := map[string]interface{}{
		"key":        key,
		"properties": properties,
	}

	forged := map[string]interface{}{
		"found": map[string]interface{}{
			"version": version,
			"entity":  entity,
		},
	}

	log.Println(forged)
	log.Println(forged["found"])
}
