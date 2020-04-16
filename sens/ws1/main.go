package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/senslabs/alpha/sens/errors"
	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/jwt"
	"github.com/senslabs/alpha/sens/logger"
	"github.com/senslabs/lambda/sens/fission/config"
	"github.com/senslabs/lambda/sens/fission/request"
	"github.com/senslabs/lambda/sens/fission/response"
	"golang.org/x/crypto/bcrypt"
)

func GetDatastoreUrl() string {
	url := os.Getenv("DATASTORE_BASE_URL")
	if url == "" {
		return "http://datastore.senslabs.me"
	}
	return url
}

func getDetail(w http.ResponseWriter, r *http.Request, category string, Id string) {
	url := fmt.Sprintf("%s/api/%s-detail-views/find", GetDatastoreUrl(), category)
	or := httpclient.HttpParams{"or": {"Id^" + Id}, "limit": {"1"}}
	var sub []map[string]interface{}
	code, err := httpclient.Get(url, or, nil, &sub)
	logger.Debugf("Code: %d, detail: %#v", code, sub)
	if err != nil {
		logger.Error(err)
		httpclient.WriteError(w, code, err)
	} else {
		generateLoginTokens(w, r, sub[0], category)
	}
}

func createAuth(body []byte) (string, error) {
	url := fmt.Sprintf("%s/api/auths/create", GetDatastoreUrl())
	code, data, err := httpclient.PostR(url, nil, nil, body)
	logger.Debugf("Code: %d, AuthId: %s", code, data)
	if err != nil {
		logger.Error(err)
		return "", err
	} else {
		return string(data), nil
	}
}

func createOrg(w http.ResponseWriter, r *http.Request, m map[string]interface{}, authId string, added bool) {
	m["AuthId"] = authId
	if b, err := json.Marshal(m); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		url := fmt.Sprintf("%s/api/orgs/create", GetDatastoreUrl())
		code, data, err := httpclient.PostR(url, nil, nil, b)
		logger.Debugf("Code: %d, Data: %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, http.StatusInternalServerError, err)
		} else if added {
			fmt.Fprintf(w, `{"OrgId":"%s"}`, data)
		} else {
			getDetail(w, r, "org", string(data))
		}
	}
}

func CreateOrg(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		authorization = r.Header.Get("authorization")
	}
	accessToken := strings.TrimPrefix(authorization, "Bearer ")

	var m map[string]interface{}
	if body, err := ioutil.ReadAll(r.Body); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusUnauthorized, err)
	} else if sub, err := jwt.VerifyToken(accessToken); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusUnauthorized, err)
	} else if err := json.Unmarshal(body, &m); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if sub["Exists"] == "auth" {
		m["Status"] = "PENDING"
		createOrg(w, r, m, sub["AuthId"].(string), false)
	} else if sub["Exists"] == "no" && m[sub["Medium"].(string)] != sub["MediumValue"] {
		logger.Error("The operation has to be done with the same Medium")
		httpclient.WriteError(w, http.StatusUnauthorized, errors.New(errors.GO_ERROR, "The operation has to be done with the same Medium"))
	} else if authId, err := createAuth(body); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		m["Status"] = "PENDING"
		createOrg(w, r, m, authId, false)
	}
}

func AddOrg(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		authorization = r.Header.Get("authorization")
	}
	var m map[string]interface{}
	accessToken := strings.TrimPrefix(authorization, "Bearer ")
	if body, err := ioutil.ReadAll(r.Body); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if sub, err := jwt.VerifyToken(accessToken); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusUnauthorized, err)
	} else if err := json.Unmarshal(body, &m); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if sub["Category"] != "auth" || m["Status"] != "APPROVED" {
		logger.Error("You are not authorized for this operation")
		httpclient.WriteError(w, http.StatusUnauthorized, errors.New(errors.GO_ERROR, "You are not authorized for this operation"))
	} else if orgsub, err := jwt.VerifyToken(m["AccessToken"]); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if orgsub["Exists"] == "auth" {
		createOrg(w, r, m, orgsub["AuthId"].(string), true)
	} else if orgsub["Exists"] == "no" && orgsub["MediumValue"] != m[orgsub["Medium"].(string)] {
		logger.Error("The operation has to be done with the same Medium")
		httpclient.WriteError(w, http.StatusUnauthorized, errors.New(errors.GO_ERROR, "The operation has to be done with the same Medium"))
	} else if authId, err := createAuth(body); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		createOrg(w, r, m, authId, true)
	}
}

func getOrgIdFromName(w http.ResponseWriter, r *http.Request, m map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/orgs/find", GetDatastoreUrl())
	or := httpclient.HttpParams{"or": {"Name^" + m["OrgName"].(string)}, "limit": {"1"}}
	code, data, err := httpclient.GetR(url, or, nil)
	logger.Debugf("Code: %d, Data: %s", code, data)
	if err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		m["OrgId"] = string(data)
	}
	return err
}

func createOp(w http.ResponseWriter, r *http.Request, m map[string]interface{}, authId string, added bool) {
	m["AuthId"] = authId
	// if err := getOrgIdFromName(w, r, m); err != nil {
	// 	logger.Error(err)
	// } else
	if b, err := json.Marshal(m); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		url := fmt.Sprintf("%s/api/ops/create", GetDatastoreUrl())
		code, data, err := httpclient.PostR(url, nil, nil, b)
		logger.Debugf("Code: %d, Data: %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, http.StatusInternalServerError, err)
		} else if added {
			fmt.Fprintf(w, `{"OpId":"%s"}`, data)
		} else {
			getDetail(w, r, "org", string(data))
		}
	}
}

func CreateOp(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		authorization = r.Header.Get("authorization")
	}

	var m map[string]interface{}
	accessToken := strings.TrimPrefix(authorization, "Bearer ")
	if body, err := ioutil.ReadAll(r.Body); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusUnauthorized, err)
	} else if sub, err := jwt.VerifyToken(accessToken); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusUnauthorized, err)
	} else if err := json.Unmarshal(body, &m); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if sub["Exists"] == "auth" {
		m["Status"] = "PENDING"
		createOp(w, r, m, sub["AuthId"].(string), false)
	} else if sub["Exists"] == "no" && m[sub["Medium"].(string)] != sub["MediumValue"] {
		logger.Error("The operation has to be done with the same Medium")
		httpclient.WriteError(w, http.StatusUnauthorized, errors.New(errors.GO_ERROR, "The operation has to be done with the same Medium"))
	} else if authId, err := createAuth(body); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		m["Status"] = "PENDING"
		createOp(w, r, m, authId, false)
	}
}

func AddOp(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		authorization = r.Header.Get("authorization")
	}
	var m map[string]interface{}
	accessToken := strings.TrimPrefix(authorization, "Bearer ")
	if body, err := ioutil.ReadAll(r.Body); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusUnauthorized, err)
	} else if sub, err := jwt.VerifyToken(accessToken); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusUnauthorized, err)
	} else if err := json.Unmarshal(body, &m); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if sub["Category"] != "org" || m["Status"] != "APPROVED" {
		logger.Error("You are not authorized for this operation")
		httpclient.WriteError(w, http.StatusUnauthorized, errors.New(errors.GO_ERROR, "You are not authorized for this operation"))
	} else if opsub, err := jwt.VerifyToken(m["AccessToken"]); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if opsub["Exists"] == "auth" {
		m["OrgId"] = sub["OrgId"]
		createOp(w, r, m, opsub["AuthId"].(string), true)
	} else if opsub["Exists"] == "no" && opsub["MediumValue"] != m[opsub["Medium"].(string)] {
		logger.Error("The operation has to be done with the same Medium")
		httpclient.WriteError(w, http.StatusUnauthorized, errors.New(errors.GO_ERROR, "The operation has to be done with the same Medium"))
	} else if authId, err := createAuth(body); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		m["OrgId"] = sub["OrgId"]
		createOp(w, r, m, authId, true)
	}
}

func createUser(w http.ResponseWriter, r *http.Request, m map[string]interface{}, authId string, added bool) {
	m["AuthId"] = authId
	// if err := getOrgIdFromName(w, r, m); err != nil {
	// 	logger.Error(err)
	// } else
	if b, err := json.Marshal(m); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		url := fmt.Sprintf("%s/api/users/create", GetDatastoreUrl())
		code, data, err := httpclient.PostR(url, nil, nil, b)
		logger.Debugf("Code: %d, Data: %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, http.StatusInternalServerError, err)
		} else if added {
			fmt.Fprintf(w, `{"UserId":"%s"}`, data)
		} else {
			getDetail(w, r, "user", string(data))
		}
	}
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		authorization = r.Header.Get("authorization")
	}

	var m map[string]interface{}
	accessToken := strings.TrimPrefix(authorization, "Bearer ")
	if body, err := ioutil.ReadAll(r.Body); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusUnauthorized, err)
	} else if sub, err := jwt.VerifyToken(accessToken); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusUnauthorized, err)
	} else if err := json.Unmarshal(body, &m); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if sub["Exists"] == "auth" {
		m["Status"] = "PENDING"
		createUser(w, r, m, sub["AuthId"].(string), false)
	} else if sub["Exists"] == "no" && m[sub["Medium"].(string)] != sub["MediumValue"] {
		logger.Error("The operation has to be done with the same Medium")
		httpclient.WriteError(w, http.StatusUnauthorized, errors.New(errors.GO_ERROR, "The operation has to be done with the same Medium"))
	} else if authId, err := createAuth(body); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		m["Status"] = "PENDING"
		createUser(w, r, m, authId, false)
	}
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		authorization = r.Header.Get("authorization")
	}
	var m map[string]interface{}
	accessToken := strings.TrimPrefix(authorization, "Bearer ")
	if body, err := ioutil.ReadAll(r.Body); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if sub, err := jwt.VerifyToken(accessToken); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusUnauthorized, err)
	} else if err := json.Unmarshal(body, &m); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if (sub["Category"] != "op" && sub["Category"] != "auth") || m["Status"] != "APPROVED" {
		logger.Error("You are not authorized for this operation")
		httpclient.WriteError(w, http.StatusUnauthorized, errors.New(errors.GO_ERROR, "You are not authorized for this operation"))
	} else if usersub, err := jwt.VerifyToken(m["AccessToken"]); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if usersub["Exists"] == "auth" {
		logger.Debugf("%#v", sub)
		m["OrgId"] = sub["OrgId"]
		createUser(w, r, m, usersub["AuthId"].(string), true)
	} else if usersub["Exists"] == "no" && usersub["MediumValue"] != m[usersub["Medium"].(string)] {
		logger.Error("The operation has to be done with the same Medium")
		httpclient.WriteError(w, http.StatusUnauthorized, errors.New(errors.GO_ERROR, "The operation has to be done with the same Medium"))
	} else if authId, err := createAuth(body); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		m["OrgId"] = sub["OrgId"]
		createUser(w, r, m, authId, true)
	}
}

///DEVICE

type Device struct {
	DeviceId   string      `json:",omitempty"`
	Name       string      `json:",omitempty"`
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
	url := fmt.Sprintf("%s%s", GetDatastoreUrl(), "/api/devices/create")
	code, data, err := httpclient.PostR(url, nil, nil, r.Body)
	logger.Debug(code, string(data))
	if err != nil {
		logger.Error(err)
		httpclient.WriteError(w, code, err)
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
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		duplicateDevice(w, r, sub["OrgId"].(string), "", REGISTERED)
	}
}

func UnregisterDevice(w http.ResponseWriter, r *http.Request) {
	duplicateDevice(w, r, "", "", UNREGISTERED)
}

func PairDevice(w http.ResponseWriter, r *http.Request) {
	var body DeviceUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		duplicateDevice(w, r, "", body.UserId, PAIRED)
	}
}

func UnpairDevice(w http.ResponseWriter, r *http.Request) {
	duplicateDevice(w, r, "", "", UNPAIRED)
}

func ListDevices(w http.ResponseWriter, r *http.Request) {
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

///ORG

func ListOrgs(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Enter ListOrgs")
	if sub, err := getAuthSubject(r); err != nil {
		httpclient.WriteUnauthorizedError(w, err)
	} else if sub["IsSens"].(bool) {
		url := fmt.Sprintf("%s/api/org-detail-views/find", GetDatastoreUrl())
		code, data, err := httpclient.GetR(url, httpclient.HttpParams{"limit": {"1000"}}, nil)
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func ListOps(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		url := fmt.Sprintf("%s/api/op-detail-views/find", GetDatastoreUrl())
		or := httpclient.HttpParams{"or": {"OrgId^" + sub["OrgId"].(string)}, "limit": {"100"}}
		code, data, err := httpclient.GetR(url, or, nil)
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		url := fmt.Sprintf("%s/api/user-detail-views/find", GetDatastoreUrl())
		or := httpclient.HttpParams{"or": {"OrgId^" + sub["OrgId"].(string)}, "limit": {"100"}}
		code, data, err := httpclient.GetR(url, or, nil)
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func getUserParams(w http.ResponseWriter, r *http.Request, orgId string) httpclient.HttpParams {
	userIds := r.URL.Query()["userId"]
	users := "UserId^" + strings.Join(userIds, "^")
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "null"
	}
	return httpclient.HttpParams{"in": {users}, "and": {"OrgId^" + orgId}, "limit": {limit}}
}

func ListUserSummary(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		params := getUserParams(w, r, sub["OrgId"].(string))
		url := fmt.Sprintf("%s/api/user-summary-views/find", GetDatastoreUrl())
		code, data, err := httpclient.GetR(url, params, nil)
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func ListUserSleeps(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		params := getUserParams(w, r, sub["OrgId"].(string))
		url := fmt.Sprintf("%s/api/user-sleep-views/find", GetDatastoreUrl())
		code, data, err := httpclient.GetR(url, params, nil)
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func ListUserMeditations(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		params := getUserParams(w, r, sub["OrgId"].(string))
		url := fmt.Sprintf("%s/api/user-meditation-views/find", GetDatastoreUrl())
		code, data, err := httpclient.GetR(url, params, nil)
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func ListActivities(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		params := getUserParams(w, r, sub["OrgId"].(string))
		params["days"] = r.URL.Query()["days"]
		url := fmt.Sprintf("%s/api/ext/activities/get", GetDatastoreUrl())
		logger.Debugf("%#v", params)
		code, data, err := httpclient.GetR(url, params, nil)
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteInternalServerError(w, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func GetSleepsSummary(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		params := getUserParams(w, r, sub["OrgId"].(string))
		url := fmt.Sprintf("%s/api/sleep-summaries/find", GetDatastoreUrl())
		logger.Debugf("%s, %#v", url, params)
		code, data, err := httpclient.GetR(url, params, nil)
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func GetMeditationsSummary(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		params := getUserParams(w, r, sub["OrgId"].(string))
		url := fmt.Sprintf("%s/api/meditation-summaries/find", GetDatastoreUrl())
		logger.Debugf("%s, %#v", url, params)
		code, data, err := httpclient.GetR(url, params, nil)
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func ListAlerts(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		params := getUserParams(w, r, sub["OrgId"].(string))
		url := fmt.Sprintf("%s/api/user-alert-views/find", GetDatastoreUrl())
		code, data, err := httpclient.GetR(url, params, nil)
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func ListKeys(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		params := httpclient.HttpParams{"limit": r.URL.Query()["limit"], "or": {"OrgId^" + sub["OrgId"].(string)}}
		url := fmt.Sprintf("%s/api/api-keys/find", GetDatastoreUrl())
		code, data, err := httpclient.GetR(url, params, nil)
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func CreateKey(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		body := make(map[string]string)
		body["ApiKey"] = fmt.Sprintf("%x", b)
		if sub, err := getAuthSubject(r); err != nil {
			logger.Error(err)
			httpclient.WriteUnauthorizedError(w, err)
		} else if keyId, err := uuid.NewRandom(); err != nil {
			logger.Error(err)
			httpclient.WriteInternalServerError(w, err)
		} else if hash, err := bcrypt.GenerateFromPassword([]byte(body["ApiKey"]), bcrypt.DefaultCost); err != nil {
			logger.Error(err)
			httpclient.WriteInternalServerError(w, err)
		} else {
			body["OrgId"] = sub["OrgId"].(string)
			body["KeyName"] = keyId.String()
			body["Key"] = string(hash)
			body["ApiKeyId"] = keyId.String()
			if b, err := json.Marshal(body); err != nil {
				logger.Error(err)
				httpclient.WriteInternalServerError(w, err)
			} else {
				url := fmt.Sprintf("%s/api/api-keys/create", GetDatastoreUrl())
				code, data, err := httpclient.PostR(url, nil, nil, b)
				logger.Debugf("%d, %s", code, data)
				if err != nil {
					logger.Error(err)
					httpclient.WriteInternalServerError(w, err)
				} else {
					fmt.Fprintf(w, "%s", b)
				}
			}
		}
	}
}

///REQUESTOTP

type AuthRequestBody struct {
	Medium      string
	MediumValue string
}

type TwilioSendOtpResponse struct {
	ServiceSid string `json:"service_sid"`
}

func VerifyAuth(w http.ResponseWriter, r *http.Request) {
	var reqBody AuthRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		params := httpclient.HttpParams{"and": {reqBody.Medium + "^" + reqBody.MediumValue, "IsSens^true"}, "limit": {"1"}}
		url := fmt.Sprintf("%s/api/auths/find", GetDatastoreUrl())
		var m []map[string]interface{}
		code, err := httpclient.Get(url, params, nil, &m)
		logger.Debugf("Code: %d, Data: %#v", code, m)
		if err != nil || len(m) == 0 {
			logger.Error(err)
			httpclient.WriteUnauthorizedError(w, err)
		} else {
			generateAuthToken(w, r, m[0], "auth")
		}
	}
}

///SESSIONS

type Session struct {
	Id        string `json:"Id"`
	UserId    string `json:"UserId"`
	Name      string `json:"Name"`
	Type      string `json:"Type"`
	StartedAt int64  `json:"StartedAt"`
	EndedAt   int64  `json:"EndedAt"`
}

type Sessions []Session

type SessionProperty struct {
	SessionId string `json:"SessionId"`
	Name      string `json:"name"`
	Value     string `json:"value"`
}

type SessionProperties []SessionProperty

type SessionRecord struct {
	UserId    string  `json:"UserId"`
	Name      string  `json:"Name"`
	Timestamp int64   `json:"Timestamp"`
	Value     float64 `json:"Value"`
}

type SessionRecords []SessionRecord

type SessionEvent struct {
	Id        string `json:"Id"`
	UserId    string `json:"UserId"`
	Name      string `json:"Name"`
	StartTime int64  `json:"StartTime"`
	EndTime   int64  `json:"EndTime"`
}

type SessionEvents []SessionEvent

type SessionSleep struct {
	Duration            int64
	RecoveryValue       int64
	RecommendedRecovery int64
	SleepTime           int64
	WakeupTime          int64
	Score               int64
	HeartRates          TimeSeriesData
	BreathRates         TimeSeriesData
	Recovery            TimeSeriesData
	Stress              TimeSeriesData
	Stages              TimeSeriesData
	AverageVitals       struct {
		HeartRate  int64
		BreathRate int64
		Stress     int64
	}
	Points struct {
		SleepQuality int64
		SleepRoutine int64
		Vitals       int64
		Restfulness  int64
	}
}

type SessionSleepData []SessionSleep

type SessionSnapshot struct {
	LastSync   int64
	Score      int64
	HeartRate  int64
	BreathRate int64
	Duration   int64
	Recovery   int64
	Id         string
	UserId     string
	StartTime  int64
	EndTime    int64
}

type SessionSnapshots []SessionSnapshot

type OperatorUser struct {
	OpId   string `json:"OpId"`
	UserId string `json:"UserId"`
}

type OperatorUsers []OperatorUser

type OrganizationUser struct {
	OrgId  string `json:"OrgId"`
	UserId string `json:"UserId"`
}

type OrganizationUsers []OrganizationUser

type TimeSeries struct {
	Time  int64
	Value interface{}
}

type TimeSeriesData []TimeSeries

type SessionsSummary struct {
	Sleeps      int64
	Meditations int64
	Alerts      int64
	Date        int64
}

func fetchSessionProperties(sessionId string, requiredSessionProperties map[string]int64) map[string]int64 {
	// Fetching session properties
	for key := range requiredSessionProperties {
		sessionPropertiesUrl := fmt.Sprintf("%v/api/session-properties/find?and=SessionId^%v&and=Name^%v&limit=1", config.GetDatastoreUrl(), sessionId, key)
		sessionPropertiesResponseData := getFromDataStore(sessionPropertiesUrl)

		var sessionPropertiesData SessionProperties
		err := json.Unmarshal(sessionPropertiesResponseData, &sessionPropertiesData)
		if err != nil {
			log.Println("Error unmarshalling response data to session properties")
		}

		if len(sessionPropertiesData) > 0 {
			sValue := sessionPropertiesData[0].Value
			var value int64
			if key == "HeartRate" || key == "BreathRate" || key == "Recovery" || key == "Score" {
				value, _ = strconv.ParseInt(sValue, 10, 64)
			} else if key == "WakeupTime" || key == "SleepTime" || key == "SunriseTime" {
				value, _ = strconv.ParseInt(sValue, 10, 64)
			} else {
				value, _ = strconv.ParseInt(sValue, 10, 64)
			}
			requiredSessionProperties[key] = value
		}
	}

	return requiredSessionProperties
}

func fetchSessionRecords(sessionUserId string, sessionStartTime int64, sessionEndTime int64, requiredSessionRecords *map[string]TimeSeriesData) {
	// Fetch records
	for key := range *requiredSessionRecords {
		sessionRecordsUrl := fmt.Sprintf("%v/api/session-records/find?and=Name^%v&and=UserId^%v&span=Timestamp^%v^%v&limit=10000000", config.GetDatastoreUrl(), key, sessionUserId, sessionStartTime, sessionEndTime)
		sessionRecordsReponseData := getFromDataStore(sessionRecordsUrl)
		var sessionRecordsData SessionRecords
		json.Unmarshal(sessionRecordsReponseData, &sessionRecordsData)

		for _, value := range sessionRecordsData {
			timestamp := value.Timestamp
			var newEvent TimeSeries
			newEvent.Time = timestamp
			if key == "HeartRate" || key == "BreathRate" || key == "Stress" || key == "Recovery" || key == "StftRatio" {
				newEvent.Value = value.Value
			} else if key == "Stage" {
				newEvent.Value = int64(value.Value)
			}
			(*requiredSessionRecords)[key] = append((*requiredSessionRecords)[key], newEvent)
		}
	}
}

func getUserSessions(r *http.Request, sessionType string, limit int64, userId *string) Sessions {
	sFrom := request.GetQueryParam(r, "from")
	var from int64
	var to int64
	if len(sFrom) != 0 {
		from, _ = strconv.ParseInt(sFrom, 10, 64)
	}
	sTo := request.GetQueryParam(r, "to")
	if len(sTo) != 0 {
		to, _ = strconv.ParseInt(sTo, 10, 64)
	}

	var userIdList []string
	if userId == nil {
		userIdList = getUserList(r)
	} else {
		userIdList = append(userIdList, *userId)
	}

	userSessionsData := make(Sessions, 0)

	for _, currentUserId := range userIdList {
		var url string
		url = fmt.Sprintf("%v/api/sessions/find?and=UserId^%v&limit=%v&and=Type^%v&column=EndedAt", config.GetDatastoreUrl(), currentUserId, limit, sessionType)
		if from != 0 && to != 0 {
			url = fmt.Sprintf("%v&span=EndedAt^%v^%v", url, from, to)
		}

		userSessionResponseData := getFromDataStore(url)
		var currentSessionData Sessions
		json.Unmarshal(userSessionResponseData, &currentSessionData)
		userSessionsData = append(userSessionsData, currentSessionData...)
	}

	return userSessionsData
}

func getSessionSnapshot(sessionId string, sessionType string) SessionSnapshot {
	sessionData := getSessionData(sessionId)

	sessionUserId := sessionData.UserId
	sessionStartTime := sessionData.StartedAt
	sessionEndTime := sessionData.EndedAt

	requiredSessionProperties := map[string]int64{
		"Recovery": 0,
		"Score":    0,
	}

	if sessionType == "Sleep" {
		requiredSessionProperties["SleepTime"] = 0
		requiredSessionProperties["WakeupTime"] = 0
	} else {
		requiredSessionProperties["Duration"] = 0
	}

	requiredSessionProperties = fetchSessionProperties(sessionId, requiredSessionProperties)

	requiredSessionRecords := map[string]TimeSeriesData{
		"HeartRate":  make(TimeSeriesData, 0),
		"BreathRate": make(TimeSeriesData, 0),
		"Recovery":   make(TimeSeriesData, 0),
		"Stress":     make(TimeSeriesData, 0),
	}

	if sessionType == "Sleep" {
		requiredSessionRecords["Stage"] = make(TimeSeriesData, 0)
	}

	fetchSessionRecords(sessionUserId, sessionStartTime, sessionEndTime, &requiredSessionRecords)

	sessionSnapshot := createSessionSnapshotData(sessionData, requiredSessionProperties, requiredSessionRecords)

	return sessionSnapshot
}

func createSessionSnapshotData(sessionData Session, requiredSessionProperties map[string]int64, requiredSessionRecords map[string]TimeSeriesData) SessionSnapshot {
	sessionSleepTime := sessionData.StartedAt
	sessionWakeupTime := sessionData.EndedAt
	sessionType := sessionData.Type

	if sessionType == "Sleep" {
		sessionSleepTime = requiredSessionProperties["SleepTime"]
		sessionWakeupTime = requiredSessionProperties["WakeupTime"]
	}
	sessionScore := requiredSessionProperties["Score"]
	sessionRecovery := requiredSessionProperties["Recovery"]
	sessionHeartRateAverage := getVitalsAverage(requiredSessionRecords["HeartRate"], sessionSleepTime, sessionWakeupTime)
	sessionBreathRateAverage := getVitalsAverage(requiredSessionRecords["BreathRate"], sessionSleepTime, sessionWakeupTime)
	var sessionSleepDuration int64
	if sessionType == "Sleep" {
		sessionSleepDuration = getTotalDurationFromStages(requiredSessionRecords["Stage"])
	} else {
		sessionSleepDuration = requiredSessionProperties["Duration"]
	}

	var sessionSnapshot SessionSnapshot
	sessionSnapshot.Duration = sessionSleepDuration
	sessionSnapshot.Score = sessionScore
	sessionSnapshot.BreathRate = sessionBreathRateAverage
	sessionSnapshot.HeartRate = sessionHeartRateAverage
	sessionSnapshot.LastSync = sessionData.EndedAt
	sessionSnapshot.Recovery = sessionRecovery
	sessionSnapshot.Id = sessionData.Id
	sessionSnapshot.UserId = sessionData.UserId
	if sessionType == "Sleep" {
		sessionSnapshot.StartTime = sessionSleepTime
		sessionSnapshot.EndTime = sessionWakeupTime
	} else {
		sessionSnapshot.StartTime = sessionData.StartedAt
		sessionSnapshot.EndTime = sessionData.EndedAt
	}

	return sessionSnapshot
}

func getVitalsAverage(data TimeSeriesData, sleepTime int64, wakeupTime int64) int64 {
	var vitalsBetweenSleepTime float64
	var vitalsBetweenSleepTimeCounter float64
	for _, vital := range data {
		if vital.Time >= sleepTime && vital.Time <= wakeupTime {
			vitalValue := vital.Value.(float64)
			vitalsBetweenSleepTime += vitalValue
			vitalsBetweenSleepTimeCounter++
		}
	}
	var averageVital int64
	if vitalsBetweenSleepTimeCounter > 0 {
		averageVital = int64(vitalsBetweenSleepTime / vitalsBetweenSleepTimeCounter)
	} else {
		averageVital = 0
	}
	return averageVital
}

func getTotalDurationFromStages(stages TimeSeriesData) int64 {
	var sessionSleepDuration int64
	var eventTimeDiff int64
	var previousTime int64
	var sleepStageCounter int64
	for _, stage := range stages {
		if eventTimeDiff == 0 && previousTime == 0 {
			previousTime = stage.Time
		} else if eventTimeDiff == 0 && previousTime != 0 {
			eventTimeDiff = stage.Time - previousTime
		}
		currentStage := stage.Value.(int64)
		if currentStage != 4 {
			sleepStageCounter += 1
		}
	}
	sessionSleepDuration = sleepStageCounter * (eventTimeDiff / 1000)
	return sessionSleepDuration
}

func getUserList(r *http.Request) []string {
	var userIdList []string

	if sub, err := getAuthSubject(r); err == nil {
		orgId := sub["OrgId"]
		opId := sub["OpId"]
		userId := sub["UserId"]

		if orgId != nil {
			// fetch users under this organization id
			orgUsers := getOrganizationUsers(orgId.(string))
			userIdList = append(userIdList, orgUsers...)
		} else if len(opId.(string)) != 0 {
			// fetch users under this operator id
			opUsers := getOperatorUsers(opId.(string))
			userIdList = append(userIdList, opUsers...)
		} else if len(userId.(string)) != 0 {
			// add userId to the userIdList
			userIdList = append(userIdList, userId.(string))
		}
	}

	return userIdList
}

func getSessionData(sessionId string) Session {
	sessionUrl := fmt.Sprintf("%v/api/sessions/%v/get", config.GetDatastoreUrl(), sessionId)

	sessionResponseData := getFromDataStore(sessionUrl)

	var sessionData Session
	err := json.Unmarshal(sessionResponseData, &sessionData)

	if err != nil {
		log.Printf("Error unmarshalling response data to sleep data : %v", err)
	}

	return sessionData
}

func getOrganizationUsers(orgId string) []string {
	var orgUsers []string
	url := fmt.Sprintf("%v/api/org-users/find?and=OrgId^%v&limit=10000", config.GetDatastoreUrl(), orgId)
	organizationUsersResponseData := getFromDataStore(url)
	var organizationUserData OrganizationUsers

	err := json.Unmarshal(organizationUsersResponseData, &organizationUserData)
	if err != nil {
		log.Println("Error fetching org users")
	}

	for _, orgUser := range organizationUserData {
		orgUsers = append(orgUsers, orgUser.UserId)
	}

	return orgUsers
}

func getOperatorUsers(orgId string) []string {
	var opUsers []string
	url := fmt.Sprintf("%v/api/op-users/find?and=OrgId^%v&limit=10000", config.GetDatastoreUrl(), orgId)
	operatorUsersResponseData := getFromDataStore(url)
	var operatorUserData OperatorUsers
	err := json.Unmarshal(operatorUsersResponseData, &operatorUserData)
	if err != nil {
		log.Println("Error fetching operator users")
		return opUsers
	}

	for _, opUser := range operatorUserData {
		opUsers = append(opUsers, opUser.UserId)
	}

	return opUsers
}

func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func getDayStart(timestamp int64) int64 {
	timestampTime := time.Unix(timestamp, 0)
	startOfDay := Bod(timestampTime)
	startOfDayUnix := startOfDay.Unix()
	return startOfDayUnix
}

func getFromDataStore(URL string) []byte {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Panicf("Error creating a new request for fetching %v", URL)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic("Error performing a request")
	}
	responseBody, _ := ioutil.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		log.Panic("Error closing the response body")
	}
	return responseBody
}

func GetSession(w http.ResponseWriter, r *http.Request) {
	// Get the sessionId from the request body
	// Fetch session details using the datastore link and type as sleep
	// Using the Start Time and End Time, fetch the following
	// 1. Fetch session properties using the sessionId
	// 2. Fetch records using the UserId, Timestamp should be between Session Start Time and Session End Time
	// 3. Fetch events using the UserId, Event Start Time should be between Session Start Time and Session End Time
	// Create Sleep Map Data
	// Return Sleep Map Data through Response
	sessionId := mux.Vars(r)["sessionId"]
	sessionData := getSessionData(sessionId)
	sessionUserId := sessionData.UserId
	sessionStartTime := sessionData.StartedAt
	sessionEndTime := sessionData.EndedAt

	requiredSessionProperties := map[string]int64{
		"Recovery":            0,
		"Score":               0,
		"SleepTime":           0,
		"WakeupTime":          0,
		"Duration":            0,
		"SleepQualityPoints":  0,
		"SleepRoutinePoints":  0,
		"VitalsPoints":        0,
		"RestfulnessPoints":   0,
		"RecommendedRecovery": 0,
	}

	requiredSessionProperties = fetchSessionProperties(sessionId, requiredSessionProperties)

	requiredSessionRecords := map[string]TimeSeriesData{
		"HeartRate":  make(TimeSeriesData, 0),
		"BreathRate": make(TimeSeriesData, 0),
		"Recovery":   make(TimeSeriesData, 0),
		"Stress":     make(TimeSeriesData, 0),
		"Stage":      make(TimeSeriesData, 0),
		"Snoring":    make(TimeSeriesData, 0),
	}

	fetchSessionRecords(sessionUserId, sessionStartTime, sessionEndTime, &requiredSessionRecords)

	sessionSleepTime := requiredSessionProperties["SleepTime"]
	sessionWakeupTime := requiredSessionProperties["WakeupTime"]
	sessionScore := requiredSessionProperties["Score"]
	sessionRecovery := requiredSessionProperties["Recovery"]
	sessionRecommendedRecovery := requiredSessionProperties["RecommendedRecovery"]
	sessionHeartRateAverage := getVitalsAverage(requiredSessionRecords["HeartRate"], sessionSleepTime, sessionWakeupTime)
	sessionBreathRateAverage := getVitalsAverage(requiredSessionRecords["BreathRate"], sessionSleepTime, sessionWakeupTime)
	sessionStressAverage := getVitalsAverage(requiredSessionRecords["Stress"], sessionSleepTime, sessionWakeupTime)
	sessionSleepDuration := getTotalDurationFromStages(requiredSessionRecords["Stage"])

	var sessionSessionData SessionSleep
	sessionSessionData.RecoveryValue = sessionRecovery
	sessionSessionData.RecommendedRecovery = sessionRecommendedRecovery
	sessionSessionData.Score = sessionScore
	sessionSessionData.SleepTime = sessionSleepTime
	sessionSessionData.WakeupTime = sessionWakeupTime
	sessionSessionData.Duration = sessionSleepDuration

	sessionSessionData.AverageVitals.HeartRate = sessionHeartRateAverage
	sessionSessionData.AverageVitals.BreathRate = sessionBreathRateAverage
	sessionSessionData.AverageVitals.Stress = sessionStressAverage

	sessionSessionData.Points.Vitals = requiredSessionProperties["VitalsPoints"]
	sessionSessionData.Points.SleepQuality = requiredSessionProperties["SleepQualityPoints"]
	sessionSessionData.Points.Restfulness = requiredSessionProperties["RestfulnessPoints"]
	sessionSessionData.Points.SleepRoutine = requiredSessionProperties["SleepRoutinePoints"]

	sessionSessionData.HeartRates = requiredSessionRecords["HeartRate"]
	sessionSessionData.BreathRates = requiredSessionRecords["BreathRate"]
	sessionSessionData.Recovery = requiredSessionRecords["Recovery"]
	sessionSessionData.Stages = requiredSessionRecords["Stage"]
	sessionSessionData.Stress = requiredSessionRecords["Stress"]

	json.NewEncoder(w).Encode(sessionSessionData)
}

func ListSessions(w http.ResponseWriter, r *http.Request) {
	// Get the sessionId from the request body
	// Fetch session details using the datastore link and type as sleep
	// A function which fetches the following for the session
	// 1. Last Synced At
	// 2. Score
	// 3. Heart Rate
	// 4. Breath Rate
	// 5. Session Duration
	// 6. Recovery Value
	//sessionId := urlQueryParams.Get("id")
	logger.Debug(r)
	sessionType := request.GetQueryParam(r, "type")
	if len(sessionType) == 0 {
		httpclient.WriteError(w, http.StatusBadRequest, errors.New(http.StatusBadRequest, "Type not passed with request"))
		return
	} else {
		var limit int64
		sLimit := request.GetQueryParam(r, "limit")
		if len(sLimit) != 0 {
			limit, _ = strconv.ParseInt(sLimit, 10, 64)
		} else {
			limit = 1
		}

		userSessionsData := getUserSessions(r, sessionType, limit, nil)

		sessionsSnapshots := make(map[string]SessionSnapshots, 0)

		for _, currentSession := range userSessionsData {
			currentSessionId := currentSession.Id
			currentUserId := currentSession.UserId
			currentSessionType := currentSession.Type
			sessionSnapshotData := getSessionSnapshot(currentSessionId, currentSessionType)
			sessionsSnapshots[currentUserId] = append(sessionsSnapshots[currentUserId], sessionSnapshotData)
		}

		w.Header().Add("Content-Type", "application/json")

		json.NewEncoder(w).Encode(sessionsSnapshots)
	}
}

func GetGeneralSummary(w http.ResponseWriter, r *http.Request) {
	// get days from the url query
	// take current date and then subtract the number of days to get the started_at date
	sDays := request.GetQueryParam(r, "days")
	if len(sDays) == 0 {
		http.Error(w, "Days not passed along with request", http.StatusBadRequest)
		return
	}
	days, _ := strconv.ParseInt(sDays, 10, 64)
	sStartDate := request.GetQueryParam(r, "start")
	var endDate int64
	if len(sStartDate) != 0 {
		endDate, _ = strconv.ParseInt(sStartDate, 10, 64)
	} else {
		endDate = time.Now().Unix()
	}
	startDate := endDate - days*3600*24

	userIdList := getUserList(r)

	generatedSummary := make(map[int64]SessionsSummary, 0)

	for _, currentUserId := range userIdList {
		url := fmt.Sprintf("%v/api/sessions/find?and=UserId^%v&span=EndedAt^%v^%v&limit=100000000", config.GetDatastoreUrl(), currentUserId, startDate, endDate)
		userSessionResponseData := getFromDataStore(url)
		var userSessionsData Sessions
		err := json.Unmarshal(userSessionResponseData, &userSessionsData)
		if err != nil {
			log.Println("Error unmarshalling session data")
		}

		for _, session := range userSessionsData {
			currentSessionType := session.Type
			sessionEndTime := session.EndedAt
			sessionKey := getDayStart(sessionEndTime)
			var currentDateSessionSummary SessionsSummary
			if sessionSummary, ok := generatedSummary[sessionKey]; ok {
				currentDateSessionSummary = sessionSummary
			} else {
				currentDateSessionSummary = SessionsSummary{}
			}
			if currentSessionType == "Sleep" {
				currentDateSessionSummary.Sleeps++
			} else if currentSessionType == "Meditation" {
				currentDateSessionSummary.Meditations++
			}
			generatedSummary[sessionKey] = currentDateSessionSummary
		}
	}

	generatedSummaryList := make([]SessionsSummary, 0)
	for key := range generatedSummary {
		currentSummary := generatedSummary[key]
		currentSummary.Date = key
		generatedSummaryList = append(generatedSummaryList, currentSummary)
	}

	w.Header().Add("Content-Type", "application/json")

	json.NewEncoder(w).Encode(generatedSummaryList)
}

func validateAndFetchQueryParameters(w http.ResponseWriter, r *http.Request) (string, *int64, *int64, *string, error) {
	sFrom := request.GetQueryParam(r, "from")
	sTo := request.GetQueryParam(r, "to")
	sessionId := request.GetPathParam(r, "id")
	property := request.GetQueryParam(r, "property")

	if len(sFrom) == 0 {
		return sessionId, nil, nil, nil, errors.New(http.StatusBadRequest, "from is missing in request")
	}
	if len(sTo) == 0 {
		return sessionId, nil, nil, nil, errors.New(http.StatusBadRequest, "to is missing in request")
	}
	if len(sessionId) == 0 {
		return sessionId, nil, nil, nil, errors.New(http.StatusBadRequest, "sessionId is missing in request")
	}
	if len(property) == 0 {
		return sessionId, nil, nil, nil, errors.New(http.StatusBadRequest, "property is missing in request")
	}

	from, _ := strconv.ParseInt(sFrom, 10, 64)
	to, _ := strconv.ParseInt(sTo, 10, 64)

	return sessionId, &from, &to, &property, nil
}

func GetSessionPropertyFunc(w http.ResponseWriter, r *http.Request) {
	sessionId, from, to, property, err := validateAndFetchQueryParameters(w, r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	timeData := map[string]TimeSeriesData{
		*property: make(TimeSeriesData, 0),
	}

	sessionData := getSessionData(sessionId)

	fetchSessionRecords(sessionData.UserId, *from, *to, &timeData)

	w.Header().Add("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(timeData)

	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err)
	}
}

func GetParameterWiseAdvancedSessionData(w http.ResponseWriter, r *http.Request) {
	os.Setenv("FILE_STORE", "fluentd")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.GetParameterWiseAdvancedSessionData")

	sessionId := request.GetQueryParam(r, "id")

	sessionData := getSessionData(sessionId)

	sessionUserId := sessionData.UserId
	sessionStartTime := sessionData.StartedAt
	sessionEndTime := sessionData.EndedAt

	requestedKey := r.URL.Query().Get("dataKey")

	requestedData := map[string]TimeSeriesData{
		requestedKey: make(TimeSeriesData, 0),
	}

	fetchSessionRecords(sessionUserId, sessionStartTime, sessionEndTime, &requestedData)

	json.NewEncoder(w).Encode(requestedData)
}

func GetCategoryWiseAdvancedSessionData(w http.ResponseWriter, r *http.Request) {
	os.Setenv("FILE_STORE", "fluentd")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.GetCategoryWiseAdvancedSessionData")

	sessionId := request.GetPathParam(r, "id")

	session := getSessionData(sessionId)
	sessionUserId := session.UserId
	sessionStartTime := session.StartedAt
	sessionEndTime := session.EndedAt

	categoriesData := map[string]interface{}{
		"Stress": map[string]TimeSeriesData{
			"Vlf":   make(TimeSeriesData, 0),
			"Hf":    make(TimeSeriesData, 0),
			"Rmssd": make(TimeSeriesData, 0),
			"Pnn50": make(TimeSeriesData, 0),
		},
		"OriginalStress": map[string]TimeSeriesData{
			"Stress": make(TimeSeriesData, 0),
		},
		"Sleep": map[string]TimeSeriesData{
			"Stage": make(TimeSeriesData, 0),
		},
		"Heart": map[string]TimeSeriesData{
			"JjPeaks":   make(TimeSeriesData, 0),
			"HeartRate": make(TimeSeriesData, 0),
			"Sdnn":      make(TimeSeriesData, 0),
			"Rmssd":     make(TimeSeriesData, 0),
			"Pnn50":     make(TimeSeriesData, 0),
			"Vlf":       make(TimeSeriesData, 0),
			"Hf":        make(TimeSeriesData, 0),
		},
		"HRV Pack": map[string]TimeSeriesData{
			"Sdnn":  make(TimeSeriesData, 0),
			"Rmssd": make(TimeSeriesData, 0),
			"Pnn50": make(TimeSeriesData, 0),
			"Vlf":   make(TimeSeriesData, 0),
			"Hf":    make(TimeSeriesData, 0),
		},
		"Respiration": map[string]TimeSeriesData{
			"ZeroCrossing": make(TimeSeriesData, 0),
			"Snoring":      make(TimeSeriesData, 0),
			"Apnea":        make(TimeSeriesData, 0),
		},
	}

	requestedKey := r.URL.Query().Get("dataKey")
	w.Header().Add("Content-Type", "application/json")
	if value, ok := categoriesData[requestedKey]; ok {
		typeOfValue := reflect.TypeOf(value)
		if typeOfValue == reflect.TypeOf(map[string]TimeSeriesData{}) {
			typeCastedValue := value.(map[string]TimeSeriesData)
			fetchSessionRecords(sessionUserId, sessionStartTime, sessionEndTime, &typeCastedValue)
		}

		json.NewEncoder(w).Encode(value)
	} else {
		log.Println("No category by that name found")
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Category not found!",
		})
	}
}

func ListUserSessions(w http.ResponseWriter, r *http.Request) {
	os.Setenv("FILE_STORE", "fluentd")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.ListUserSessions")

	userId := request.GetPathParam(r, "id")
	sessionType := request.GetQueryParam(r, "type")
	if len(sessionType) == 0 {
		httpclient.WriteError(w, http.StatusBadRequest, errors.New(http.StatusBadRequest, "Type not passed with request"))
	} else {
		var limit int64
		sLimit := request.GetQueryParam(r, "limit")
		if len(sLimit) != 0 {
			limit, _ = strconv.ParseInt(sLimit, 10, 64)
		} else {
			limit = 1
		}
		userSessionsData := getUserSessions(r, sessionType, limit, &userId)

		userSessionsSnapshots := make(SessionSnapshots, 0)

		for _, currentSession := range userSessionsData {
			currentSessionId := currentSession.Id
			currentSessionType := currentSession.Type
			sessionSnapshotData := getSessionSnapshot(currentSessionId, currentSessionType)
			userSessionsSnapshots = append(userSessionsSnapshots, sessionSnapshotData)
		}

		w.Header().Add("Content-Type", "application/json")

		json.NewEncoder(w).Encode(userSessionsSnapshots)
	}
}

///TOKEN

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		authorization = r.Header.Get("authorization")
	}
	token := strings.TrimPrefix(authorization, "Bearer ")
	if authorization == "" {
		w.WriteHeader(http.StatusUnauthorized)
	} else if token, err := jwt.RefreshAccessToken(token); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		fmt.Fprintf(w, `{"AccessToken":"%s"}`, token)
	}
}

func getAuthSubject(r *http.Request) (map[string]interface{}, error) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		authorization = r.Header.Get("authorization")
	}
	accessToken := strings.TrimPrefix(authorization, "Bearer ")
	if accessToken == "" {
		logger.Debug("Access Token Empty")
		//Specific handling to understand that if the client has sent Acess Token or Not
		return nil, nil
	}
	if sub, err := jwt.VerifyToken(accessToken); err != nil {
		logger.Error(err)
		return nil, err
	} else {
		logger.Debugf("%#v", sub)
		return sub, nil
	}
}

const (
	OPEN = iota
	KEY
	SENS
)

func IsSens(w http.ResponseWriter, r *http.Request) (bool, error) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		return false, err
	} else if sub != nil && sub["IsSens"] != nil {
		logger.Debugf("%#v", sub)
		return sub["IsSens"].(bool), nil
	} else {
		return false, err
	}
}

type org struct {
	OrgName string
	OrgId   string
}

func ChangeRoleToOp(w http.ResponseWriter, r *http.Request) {
	var o org
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		sub["OrgId"] = o.OrgId
		sub["OrgName"] = o.OrgName
		generateLoginTokens(w, r, sub, "auth")
	}
}

func ChangeRoleToAuth(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		delete(sub, "OrgId")
		delete(sub, "OrgName")
		generateLoginTokens(w, r, sub, "auth")
	}
}

///VERIFYOTP

type VerifyRequestBody struct {
	Medium      string
	MediumValue string
	SessionId   string
	Otp         string
	Category    string
}

type TwilioVerifyOtpResponse struct {
	Status string `json:"status"`
}

func generateTemporaryToken(w http.ResponseWriter, r *http.Request, reqBody VerifyRequestBody) {
	sub := map[string]interface{}{
		"Medium":      reqBody.Medium,
		"MediumValue": reqBody.MediumValue,
		"Exists":      "no",
		"Category":    strings.ToLower(reqBody.Category),
	}
	if token, err := jwt.GenerateTemporaryToken(sub); err != nil {
		logger.Error("HTTP response code, message:", http.StatusInternalServerError, err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		logger.Debug("IMPORTANT :- This should not be logged. TOKEN: ", token)
		fmt.Fprintf(w, `{"AccessToken": "%s"}`, token)
	}
}

func generateAuthToken(w http.ResponseWriter, r *http.Request, sub map[string]interface{}, category string) {
	sub["Category"] = category
	sub["Exists"] = "auth"
	if token, err := jwt.GenerateTemporaryToken(sub); err != nil {
		logger.Error("HTTP response code, message:", http.StatusInternalServerError, err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		logger.Debug("IMPORTANT :- This should not be logged. TOKEN: ", token)
		fmt.Fprintf(w, `{"AccessToken": "%s"}`, token)
	}
}

func generateLoginTokens(w http.ResponseWriter, r *http.Request, sub map[string]interface{}, category string) {
	sub["Exists"] = "yes"
	sub["Category"] = category
	expiry := 7200 * time.Second
	if accessToken, err := jwt.GenerateAccessToken(sub, expiry); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if refreshToken, err := jwt.GenerateRefreshToken(sub); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		fmt.Fprintf(w, `{"AccessToken":"%s", "RefreshToken":"%s", "Expiry":%d}`, accessToken, refreshToken, expiry)
	}
}

func buildToken(w http.ResponseWriter, r *http.Request, reqBody VerifyRequestBody) {
	category := strings.ToLower(reqBody.Category)
	or := httpclient.HttpParams{"or": {reqBody.Medium + "^" + reqBody.MediumValue}, "limit": {"1"}}
	logger.Debugf("OR: %#v", or)
	url := fmt.Sprintf("%s/api/auths/find", GetDatastoreUrl())
	code, auths, err := httpclient.GetR(url, or, nil)
	logger.Debugf("%d, %s", code, auths)
	var subs []map[string]interface{}
	if err != nil || code != http.StatusOK {
		logger.Error("HTTP response code, message:", code, err)
		httpclient.WriteError(w, code, err)
	} else if err := json.Unmarshal(auths, &subs); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if len(subs) == 0 {
		generateTemporaryToken(w, r, reqBody) //return
	} else {
		url := fmt.Sprintf("%s/api/%s-detail-views/find", GetDatastoreUrl(), category)
		or := httpclient.HttpParams{"or": {"AuthId^" + subs[0]["AuthId"].(string)}, "limit": {"1"}}
		code, data, err := httpclient.GetR(url, or, nil)
		logger.Debugf("%d, %s", code, data)
		var detail []map[string]interface{}
		if err != nil || code != http.StatusOK {
			logger.Error("HTTP response code, message:", code, err)
			httpclient.WriteError(w, code, err)
		} else if err := json.Unmarshal(data, &detail); err != nil {
			logger.Error(err)
			httpclient.WriteError(w, http.StatusInternalServerError, err)
		} else if len(detail) == 0 {
			generateAuthToken(w, r, subs[0], category) //return
		} else {
			generateLoginTokens(w, r, detail[0], category)
		}
	}
}

func VerifyOtp(w http.ResponseWriter, r *http.Request) {
	categories := map[string]bool{"org": true, "op": true, "user": true, "auth": true}
	var reqBody VerifyRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if categories[strings.ToLower(reqBody.Category)] { // VERIFY OTP FROM TWILLIO
		body := fmt.Sprintf("To=%s&Code=%s", url.QueryEscape(reqBody.MediumValue), reqBody.Otp)
		url := fmt.Sprintf("https://verify.twilio.com/v2/Services/%s/VerificationCheck", reqBody.SessionId)
		var twilioResponse TwilioVerifyOtpResponse
		code, data, err := httpclient.PostR(url, nil, httpclient.HttpParams{
			"Authorization": {"Basic QUM2MmFlOWU0N2I2MTI2M2YyZDQwYzdjYjhjMzMyNzU4OTo4MTg4MGNhMTBmMjMxMGUxNjdlZGI1YTRmZGVjMDUxMg=="},
			"Content-Type":  {"application/x-www-form-urlencoded"},
		}, []byte(body))
		if code == http.StatusOK {
			err = json.Unmarshal(data, &twilioResponse)
		}
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, http.StatusInternalServerError, err)
		} else if code != http.StatusOK || twilioResponse.Status != "approved" {
			logger.Error("Twilio confrmation code entered does not seem to be correct, ", code, ", ", twilioResponse.Status)
			httpclient.WriteError(w, code, err)
		} else { //BUILD AUTH
			buildToken(w, r, reqBody)
		}
	} else {
		httpclient.WriteError(w, http.StatusBadRequest, errors.New(errors.GO_ERROR, "Category is not provided"))
	}
}

func main() {}
