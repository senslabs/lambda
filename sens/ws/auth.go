package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/senslabs/alpha/sens/errors"
	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/jwt"
	"github.com/senslabs/alpha/sens/logger"
)

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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.CreateOrg")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.AddOrg")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.CreateOp")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.AddOp")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.CreateUser")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.AddUser")
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

func main() {}
