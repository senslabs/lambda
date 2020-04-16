package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
	"golang.org/x/crypto/bcrypt"
)

func ListOrgs(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListOrgs")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListOps")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListUsers")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListUserSummary")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListUserSleeps")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListUserMeditations")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListActivities")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.GetSleepsSummary")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.GetMeditationsSummary")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListAlerts")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListKeys")
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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.CreateKey")
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
