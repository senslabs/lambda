package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/senslabs/alpha/sens/errors"
	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
)

type Device struct {
	DeviceId   string      `json:",omitempty"`
	DeviceName string      `json:",omitempty"`
	OrgId      string      `json:",omitempty"`
	UserId     string      `json:",omitempty"`
	CreatedAt  int64       `json:",omitempty"`
	Status     string      `json:",omitempty"`
	Properties interface{} `json:",omitempty"`
}

func duplicateDevice(w http.ResponseWriter, r *http.Request, orgId string, userId string, status string) error {
	if orgId == "" && userId == "" && status == "" {
		return httpclient.WriteError(w, http.StatusBadRequest, errors.New(http.StatusBadRequest, "No change in data"))
	}
	deviceId := r.URL.Query().Get("deviceId")
	var devices []Device
	url := fmt.Sprintf("%s%s", GetDatastoreUrl(), "/api/devices/find")
	and := httpclient.HttpParams{"and": {"DeviceId^" + deviceId}, "column": {"CreatedAt"}, "limit": {"1"}}
	code, err := httpclient.Get(url, and, nil, &devices)
	if len(devices) == 0 {
		return httpclient.WriteError(w, http.StatusBadRequest, errors.New(errors.DB_ERROR, "No devices found"))
	} else if status == devices[0].Status {
		return httpclient.WriteUnauthorizedError(w, errors.New(errors.DB_ERROR, "Wrong status"))
	} else if status != REGISTERED && devices[0].OrgId != orgId {
		return httpclient.WriteUnauthorizedError(w, errors.New(errors.DB_ERROR, "Wrong organization"))
	} else {
		device := devices[0]
		logger.Debugf("%d, %#v", code, device)
		if err != nil {
			return httpclient.WriteError(w, code, err)
		} else {
			device.CreatedAt = time.Now().Unix()
			if status != "" {
				device.Status = status
			}
			if orgId != "" {
				device.OrgId = orgId
			}
			if userId != "" {
				device.UserId = userId
			}
			url := fmt.Sprintf("%s%s", GetDatastoreUrl(), "/api/devices/create")
			if body, err := json.Marshal(device); err != nil {
				return httpclient.WriteError(w, code, err)
			} else if code, _, err := httpclient.PostR(url, nil, nil, body); err != nil {
				return httpclient.WriteError(w, code, err)
			} else {
				w.WriteHeader(code)
			}
		}
	}
	return nil
}

func CreateDevice(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.CreateDevice")
	if isSens, err := IsSens(w, r); err == nil && isSens {
		url := fmt.Sprintf("%s%s", GetDatastoreUrl(), "/api/devices/create")
		code, data, err := httpclient.PostR(url, nil, nil, r.Body)
		defer r.Body.Close()
		logger.Debug(code, string(data))
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintln(w, string(data))
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.RegisterDevice")
	var body DeviceUpdateBody
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else if isSens, err := IsSens(w, r); err == nil && isSens {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			defer r.Body.Close()
			logger.Error(err)
			httpclient.WriteInternalServerError(w, err)
		} else {
			duplicateDevice(w, r, body.OrgId, "", REGISTERED)
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func UnregisterDevice(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.UnregisterDevice")
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		duplicateDevice(w, r, "", sub["OrgId"].(string), UNREGISTERED)
	}
}

func PairDevice(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.PairDevice")
	var body DeviceUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		defer r.Body.Close()
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		duplicateDevice(w, r, sub["OrgId"].(string), body.UserId, PAIRED)
	}
}

func UnpairDevice(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.UnpairDevice")
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		duplicateDevice(w, r, sub["OrgId"].(string), "", UNPAIRED)
	}
}

func ListDevices(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListDevices")
	if sub, err := getAuthSubject(r); err != nil {
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		url := fmt.Sprintf("%s/api/device-views/find", GetDatastoreUrl())
		or := httpclient.HttpParams{"or": {"OrgId^" + sub["OrgId"].(string)}, "limit": r.URL.Query()["limit"]}
		code, data, err := httpclient.GetR(url, or, nil)
		logger.Debugf("%d, %#v", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

// devices/properties/add
func AddDeviceProperties(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.AddDeviceProperties")
	if isSens, err := IsSens(w, r); err == nil && isSens {
		url := fmt.Sprintf("%s%s", GetDatastoreUrl(), "/api/device-properties/batch/create")
		code, data, err := httpclient.PostR(url, nil, nil, r.Body)
		defer r.Body.Close()
		logger.Debug(code, string(data))
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintln(w, string(data))
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

// devices/{id}/properties/get
func GetDeviceProperties(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.GetDeviceProperties")
	if isSens, err := IsSens(w, r); err == nil && isSens {
		id := r.Header.Get("X-Fission-Params-Id")
		url := fmt.Sprintf("%s%s", GetDatastoreUrl(), "/api/device-properties/find")
		code, data, err := httpclient.GetR(url, nil, httpclient.HttpParams{"DeviceId": {id}, "limit": {"null"}})
		logger.Debug(code, string(data))
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintln(w, string(data))
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}
