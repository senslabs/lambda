package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/senslabs/alpha/sens/errors"
	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
	"github.com/senslabs/lambda/sens/fission/response"
)

func duplicateDevice(w http.ResponseWriter, r *http.Request, orgId string, userId string, status string) error {
	logger.InitLogger("")

	if orgId == "" && userId == "" && status == "" {
		return httpclient.WriteError(w, http.StatusBadRequest, errors.New(http.StatusBadRequest, "No change in data"))
	}
	deviceId := mux.Vars(r)["deviceId"]
	var device map[string]interface{}
	url := fmt.Sprintf("%s%s", config.getDatastoreUrl(), "/api/devices/find")
	and := httpclient.HttpParams{"DeviceId": {deviceId}}
	code, err := httpclient.Get(url, and, nil, &device)
	logger.Debugf("%d, %#v", code, device)
	if err != nil {
		return httpclient.WriteError(w, code, err)
	} else {
		delete(device, "Id")
		device["CreatedAt"] = time.Now()
		if status != "" {
			device["Status"] = status
		}
		if orgId != "" {
			device["OrgId"] = orgId
		}
		if userId != "" {
			device["UserId"] = userId
		}
		url := fmt.Sprintf("%s%s", config.getDatastoreUrl(), "/api/devices/create")
		code, _, err := httpclient.PostR(url, and, nil, &device)
		if err != nil {
			return httpclient.WriteError(w, code, err)
		}
	}
	return nil
}

func CreateDevice(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.CreateDevice")
	url := fmt.Sprintf("%s%s", config.getDatastoreUrl(), "/api/devices/create")
	code, data, err := httpclient.PostR(url, nil, nil, r.Body)
	logger.Debug(code, string(data))
	if err != nil {
		logger.Error(err)
		response.WriteError(w, code, err)
	} else {
		fmt.Fprintln(w, string(data))
	}
}

type DeviceUpdateBody struct {
	DeviceId string
	OrgId    string
	UserId   string
}

const (
	REGISTERED   = "REGISTERED"
	UNREGISTERED = "UNREGISTERED"
	PAIRED       = "PAIRED"
	UNPAIRED     = "UNPAIRED"
)

func RegisterDevice(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.RegisterDevice")
	var body DeviceUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		duplicateDevice(w, r, body.OrgId, body.UserId, REGISTERED)
	}
}

func UnregisterDevice(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.UnregisterDevice")
	duplicateDevice(w, r, "", "", UNREGISTERED)
}

func PairDevice(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.PairDevice")
	var body DeviceUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		duplicateDevice(w, r, body.OrgId, body.UserId, PAIRED)
	}
}

func UnpairDevice(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.UnpairDevice")
	var body DeviceUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		duplicateDevice(w, r, body.OrgId, body.UserId, UNPAIRED)
	}
}

func main() {}
