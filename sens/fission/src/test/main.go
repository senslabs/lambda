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

func getDatastoreUrl() string {
	baseUrl := os.Getenv("DATASTORE_BASE_URL")
	if baseUrl == "" {
		return "http://datastore.senslabs.me"
	}
	return baseUrl
}

func TestDatastore(w http.ResponseWriter, r *http.Request) {
	logger.InitConsoleLogger()
	var d Data
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		fmt.Fprintln(w, err)
	} else {
		baseUrl := getDatastoreUrl()
		url := fmt.Sprintf("%s%s", baseUrl, d.Path)
		logger.Debugf("%#v", d)
		if d.Method == "GET" {
			code, data, err := httpclient.GetR(url, nil, nil)
			logger.Debug(code, data, err)
			fmt.Fprintln(w, code, ",", string(data), "Error:", err)
		} else if b, err := json.Marshal(d.Body); err != nil {
			fmt.Fprintln(w, "Error:", err)
		} else {
			code, data, err := httpclient.PostR(url, nil, nil, b)
			fmt.Fprintln(w, "Http status code:", code, string(data), "Error:", err)
		}
	}
}
