package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
)

type Data struct {
	Method string
	Path   string
	Body   interface{}
}

func GetDatastoreUrl() string {
	baseUrl := os.Getenv("DATASTORE_BASE_URL")
	if baseUrl == "" {
		return "http://datastore.senslabs.me"
	}
	return baseUrl
}

func TestDatastore(w http.ResponseWriter, r *http.Request) {
	var d Data
	baseUrl := GetDatastoreUrl()
	url := fmt.Sprintf("%s%s", baseUrl, d.Path)
	json.NewDecoder(r.Body).Decode(&d)
	if d.Method == "GET" {
		code, data, err := httpclient.GetR(url, nil, nil)
		logger.Debug(code, data, err)
		fmt.Fprintln(w, code, ",", data, "Error:", err)
	} else if b, err := json.Marshal(d.Body); err != nil {
		fmt.Fprintln(w, "Error:", err)
	} else {
		code, data, err := httpclient.PostR(url, nil, nil, b)
		fmt.Fprintln(w, code, ",", data, "Error:", err)
	}
}
