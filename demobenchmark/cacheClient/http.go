package cacheClient

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

const datastoreURL string = "https://datastore.googleapis.com/v1/projects/"
const currentToken string = "ya29.GlsbB-di5INRPPru0kxDyUfbaCtYYMpkNqSVvoL9_0u9-DibsK14XPAaVeijLYfGzFPBEGZ0BKIzl1zGNButk6vMZYWbW32aSt39q0r77QzVZFxkwRU1SKKzMvCQ"
const projectid = "central-binder-241522"

type httpClient struct {
	*http.Client
	server string
}

func (c *httpClient) get(key string) bool {
	id, e := strconv.ParseInt(key, 10, 64)
	if e != nil {
		log.Fatalln(e)
	}

	requestBody := map[string]interface{}{
		"keys": map[string]interface{}{
			"path": map[string]interface{}{
				"kind": "Book",
				"id":   id,
			},
		},
	}

	bytesRepresentation, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(c.server+projectid+":lookup"+"?access_token="+currentToken,
		"application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	log.Println(resp)
	log.Println(result)

	found := false
	if resp.StatusCode == http.StatusOK {
		if result["found"] != nil {
			log.Println("Entity Found: ", result["found"])
			found = true
		}
		if result["missing"] != nil {
			log.Println("Entity Not Found: ", result["missing"])
		}
	} else {
		log.Println("WARNING: ", resp.StatusCode, resp.Body)
	}

	return found
}

func (c *httpClient) set(key string, value BookData) {
	transRequest := map[string]interface{}{
		"transactionOptions": map[string]interface{}{
			"readWrite": map[string]interface{}{},
		},
	}

	bytesRepresentation, err := json.Marshal(transRequest)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(datastoreURL+projectid+":beginTransaction"+"?access_token="+currentToken,
		"application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	var beginTransRespon map[string]string
	json.NewDecoder(resp.Body).Decode(&beginTransRespon)
	log.Println("BeginTrans Response: ", resp)

	var trans string
	if resp.StatusCode == http.StatusOK {
		trans = beginTransRespon["transaction"]
	}

	ID, err := strconv.ParseInt(value.ISBN, 10, 64)
	if err != nil {
		log.Fatalln(err)
	}

	tmp := &Book{
		ID:            ID,
		Title:         value.BookTitle,
		Author:        value.BookAuthor,
		PublishedDate: value.Year,
		ImageURL:      value.ImageM,
		//Description:   r.FormValue("description"),
		CreatedBy:   "Gopher",
		CreatedByID: "000",
	}

	//commit
	commitMessage := map[string]interface{}{
		"transaction": trans,
		"mutations": map[string]map[string]interface{}{
			"upsert": {
				"key": map[string]interface{}{
					"path": map[string]interface{}{
						"id":   tmp.ID,
						"kind": "Book",
					},
				},
				"properties": map[string]interface{}{
					"PublishedDate": map[string]interface{}{
						"stringValue": tmp.PublishedDate,
					},
					"ImageURL": map[string]interface{}{
						"stringValue": tmp.ImageURL,
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

	resp, err = http.Post(c.server+projectid+":commit"+"?access_token="+currentToken,
		"application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	log.Println("Response: ", resp)
	log.Println(result)
}

func (c *httpClient) Run(cmd *Cmd) {
	if cmd.Name == "get" {
		cmd.Res = c.get(cmd.Key)
		return
	}
	if cmd.Name == "set" {
		c.set(cmd.Key, cmd.Value)
		return
	}
	panic("unknown cmd name " + cmd.Name)
}

func newHTTPClient(server string) *httpClient {
	client := &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 1}}
	return &httpClient{client, "http://" + server + ":14350/cache/"}
}

