package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const baseURL string = "https://datastore.googleapis.com/v1/projects/"

type cacheHandler struct {
	*Server
}

func (h *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	paras := strings.Split(r.URL.EscapedPath(), "/")[2]
	if len(paras) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m := r.Method
	if m == http.MethodPut {
		b, _ := ioutil.ReadAll(r.Body)
		key := paras
		addr, e := h.ShouldProcess(key)
		if !e {
			log.Println("SET: Wrong host, Discrad key :", addr)
		} else {
			if len(b) != 0 {
				e := h.Set(key, b)
				if e != nil {
					log.Println(e)
					w.WriteHeader(http.StatusInternalServerError)
				}
				log.Println(addr, " accepts key:", key)
			}
		}
		return
	}

	//POST
	projectid := strings.Split(paras, ":")[0]
	method := strings.Split(paras, ":")[1]
	token := r.URL.RawQuery[strings.LastIndex(r.URL.RawQuery, "=")+1:]
	url := baseURL + paras + "?" + r.URL.RawQuery
	//token := r.Header.Get("Authorization")[0]
	//log.Println(method)
	//log.Println(token)

	if len(projectid) == 0 || len(method) == 0 || len(token) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing argument."))
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer r.Body.Close()

	if m == http.MethodPost {
		var req map[string]interface{}
		json.Unmarshal(bodyBytes, &req)

		if len(req) != 0 {
			if method == "lookup" {
				tmpk := req["keys"].(map[string]interface{})
				tmpp := tmpk["path"].(map[string]interface{})
				kind := tmpp["kind"].(string)
				idtmp := tmpp["id"].(float64)
				id := strconv.Itoa(int(idtmp))
				key := projectid + kind + id

				addr, e := h.ShouldProcess(key)
				if !e {
					log.Println("Transfer key query to :", addr)
					data, code := h.DoTransfer(paras+"?"+r.URL.RawQuery, addr, bodyBytes)
					w.Write(data)
					w.WriteHeader(code)
					return
				}

				b, error := h.Get(key)
				if error != nil {
					log.Println(e)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(error.Error()))
				}

				w.Header().Set("Content-Type", "application/json")
				if b == nil {
					data, code := h.DoFetch(projectid, key, token, req)

					w.Write(data)
					if code != http.StatusOK {
						w.WriteHeader(code)
					}
				} else {
					w.Write(b)
				}
			} else if method == "commit" {
				mutations := req["mutations"].(map[string]interface{})

				if mutations["upsert"] == nil || len(mutations) > 1 {
					h.DoUpdate(url, bodyBytes)
					data, code := h.DoUpdate(url, bodyBytes)
					w.Write(data)
					w.WriteHeader(code)
					return
				}

				data, code := h.DoUpdate(url, bodyBytes)
				w.Write(data)
				if code != http.StatusOK {
					w.WriteHeader(code)
				} else {
					//set to cache
					key, found := h.ConvertTofound(req, projectid, data)

					addr, e := h.ShouldProcess(key)
					if !e {
						log.Println("Transfer key value pair to :", addr)
						h.DoAssign(key, addr, found)
					} else {
						h.Set(key, found)
					}
				}
			} else {
				h.DoUpdate(url, bodyBytes)
				data, code := h.DoUpdate(url, bodyBytes)
				w.Write(data)
				w.WriteHeader(code)
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing payload."))
		}
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (h *cacheHandler) DoFetch(projectid, key, token string, message map[string]interface{}) ([]byte, int) {
	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(baseURL+projectid+":lookup"+"?access_token="+token,
		"application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	var result map[string]interface{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	//json.NewDecoder(bodyBytes).Decode(&result)
	//resp.body ([]byte)-> map

	json.Unmarshal(bodyBytes, &result)

	log.Println("Retrieved from Datastore :", resp)

	if resp.StatusCode == http.StatusOK {
		if result["found"] != nil {
			log.Println("Entity Found: ", result["found"])
		}
		if result["missing"] != nil {
			log.Println("Entity Not Found: ", result["missing"])
		}
		log.Println(result)

		h.Set(key, bodyBytes)
	} else {
		log.Println("WARNING: ", resp.StatusCode, resp.Body)
	}

	return bodyBytes, resp.StatusCode
}

func (h *cacheHandler) DoUpdate(url string, body []byte) ([]byte, int) {
	resp, err := http.Post(url,
		"application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalln(err)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	return bodyBytes, resp.StatusCode
}

func (h *cacheHandler) ConvertTofound(req map[string]interface{}, projectid string, resp []byte) (string, []byte) {
	mutations := req["mutations"].(map[string]interface{})

	upsert := mutations["upsert"].(map[string]interface{})
	properties := upsert["properties"].(map[string]interface{})
	key := upsert["key"].(map[string]interface{})
	key["partitionId"] = map[string]string{"projectId": projectid}

	path := key["path"].(map[string]interface{})
	kind := path["kind"].(string)
	idtmp := path["id"].(float64)
	id := strconv.Itoa(int(idtmp))

	extractkey := projectid + kind + id

	var temp map[string]interface{}
	json.Unmarshal(resp, &temp)
	mutationRes := temp["mutationResults"].([]interface{})
	var t interface{}
	for _, res := range mutationRes {
		t = res
	}
	t2 := t.(map[string]interface{})
	version := t2["version"]

	entity := map[string]interface{}{
		"key":        key,
		"properties": properties,
	}

	found := map[string][]interface{}{
		"found": {
			map[string]interface{}{
				"version": version,
				"entity":  entity,
			},
		},
	}

	res, err := json.Marshal(found)
	if err != nil {
		log.Fatalln("convert :", err)
	}
	return extractkey, res
}

func (h *cacheHandler) DoTransfer(url, addr string, cnt []byte) ([]byte, int) {

	resp, err := http.Post("http://"+addr+":14350/cache/"+url,
		"application/json", bytes.NewBuffer(cnt))
	if err != nil {
		log.Fatalln(err)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	return bodyBytes, resp.StatusCode

}

func (h *cacheHandler) DoAssign(key, addr string, value []byte) ([]byte, int) {
	c := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, "http://"+addr+":14350/cache/"+key,
		bytes.NewBuffer(value))
	if err != nil {
		log.Println(err)
		log.Println(http.MethodPut, "http://"+addr+":14350/cache/"+key)
		log.Println(req)
	}

	resp, err := c.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	return bodyBytes, resp.StatusCode
}

func (s *Server) cacheHandler() http.Handler {
	return &cacheHandler{s}
}
