package gcloudfn

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/senslabs/alpha/sens/errors"
	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
	"github.com/senslabs/alpha/sens/types"
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
		url := fmt.Sprintf("%s/api/org-views/find", GetDatastoreUrl())
		code, data, err := httpclient.GetR(url, httpclient.HttpParams{"limit": {"null"}}, nil)
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
		url := fmt.Sprintf("%s/api/op-views/find", GetDatastoreUrl())
		or := httpclient.HttpParams{"or": {"OrgId^" + sub["OrgId"].(string)}, "limit": {"null"}}
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
		url := fmt.Sprintf("%s/api/user-views/find", GetDatastoreUrl())
		or := httpclient.HttpParams{"or": {"OrgId^" + sub["OrgId"].(string)}, "limit": {"null"}}
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

func GetUserParams(w http.ResponseWriter, r *http.Request, orgId string) httpclient.HttpParams {
	userIds := r.URL.Query()["userId"]
	users := "UserId^" + strings.Join(userIds, "^")
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "null"
	}
	return httpclient.HttpParams{"in": {users}, "and": {"OrgId^" + orgId}, "limit": {limit}}
}

// org-activity-summary-views; was /api/users/summary
func ListUserSummary(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListUserSummary")
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		params := GetUserParams(w, r, sub["OrgId"].(string))
		url := fmt.Sprintf("%s/api/org-activity-summary-views/find", GetDatastoreUrl())
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

// org/sleeps/list; was user/sleeps/list
func ListUserSleeps(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListUserSleeps")
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		params := GetUserParams(w, r, sub["OrgId"].(string))
		params["and"] = []string{"SessionType^Sleep"}
		url := fmt.Sprintf("%s/api/org-session-info-views/find", GetDatastoreUrl())
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
		params := GetUserParams(w, r, sub["OrgId"].(string))
		params["and"] = []string{"SessionType^Meditation"}
		url := fmt.Sprintf("%s/api/org-session-info-views/find", GetDatastoreUrl())
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
		params := httpclient.HttpParams{"days": r.URL.Query()["days"], "and": {"OrgId^" + sub["OrgId"].(string)}}
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

// org/sleeps/summary; was sleeps/summary/get
func GetSleepsSummary(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.GetSleepsSummary")
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		params := GetUserParams(w, r, sub["OrgId"].(string))
		url := fmt.Sprintf("%s/api/org-sleep-views/find", GetDatastoreUrl())
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

// org/meditations/summary; was meditations/summary/get
func GetMeditationsSummary(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.GetMeditationsSummary")
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		params := GetUserParams(w, r, sub["OrgId"].(string))
		url := fmt.Sprintf("%s/api/org-meditation-views/find", GetDatastoreUrl())
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
		params := GetUserParams(w, r, sub["OrgId"].(string))
		url := fmt.Sprintf("%s/api/org-alert-views/find", GetDatastoreUrl())
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

// api/org/
func GetQuarterActivityCounts(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.GetQuarterActivityCount")
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		params := httpclient.HttpParams{"and": {"OrgId^" + sub["OrgId"].(string)}, "limit": {"null"}}
		url := fmt.Sprintf("%s/api/org-quarter-usage-views/find", GetDatastoreUrl())
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

type Property struct {
	UserId string
	Key    string
	Value  string
}

func UpdateUserProperties(w http.ResponseWriter, r *http.Request) {
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		var properties []Property
		userId := r.Header.Get("X-Fission-Params-Id")
		m := types.UnmarshalMap(r.Body)
		for k, v := range m {
			properties = append(properties, Property{userId, k, v.(string)})
		}

		url := fmt.Sprintf("%s/api/user-properties/batch/create", GetDatastoreUrl())
		code, data, err := httpclient.PostR(url, nil, nil, types.Marshal(properties))
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func UpdateOpProperties(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		var properties []Property
		opId := sub["OpId"].(string)
		m := types.UnmarshalMap(r.Body)
		for k, v := range m {
			properties = append(properties, Property{opId, k, v.(string)})
		}

		url := fmt.Sprintf("%s/api/op-properties/batch/create", GetDatastoreUrl())
		code, data, err := httpclient.PostR(url, nil, nil, types.Marshal(properties))
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func UpdateOrgProperties(w http.ResponseWriter, r *http.Request) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		var properties []Property
		orgId := sub["OrgId"].(string)
		m := types.UnmarshalMap(r.Body)
		for k, v := range m {
			properties = append(properties, Property{orgId, k, v.(string)})
		}

		url := fmt.Sprintf("%s/api/org-properties/batch/create", GetDatastoreUrl())
		code, data, err := httpclient.PostR(url, nil, nil, types.Marshal(properties))
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func ListUserProperties(w http.ResponseWriter, r *http.Request) {
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		userId := r.Header.Get("X-Fission-Params-Id")
		params := httpclient.HttpParams{"and": {"UserId^" + userId}, "limit": {"null"}}
		url := fmt.Sprintf("%s/api/user-properties/find", GetDatastoreUrl())

		var properties []Property
		code, err := httpclient.Get(url, params, nil, &properties)
		logger.Debugf("%d, %#v", code, properties)
		errors.Pie(err)

		response := map[string]interface{}{}
		for _, p := range properties {
			response[p.Key] = p.Value
		}
		types.MarshalInto(response, w)
	}
}

func ListOrgProperties(w http.ResponseWriter, r *http.Request) {
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		orgId := r.Header.Get("X-Fission-Params-Id")
		params := httpclient.HttpParams{"and": {"OrgId^" + orgId}, "limit": {"null"}}
		url := fmt.Sprintf("%s/api/org-properties/find", GetDatastoreUrl())

		var properties []Property
		code, err := httpclient.Get(url, params, nil, &properties)
		logger.Debugf("%d, %#v", code, properties)
		errors.Pie(err)

		response := map[string]interface{}{}
		for _, p := range properties {
			response[p.Key] = p.Value
		}
		types.MarshalInto(response, w)
	}
}

func ListOpProperties(w http.ResponseWriter, r *http.Request) {
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		opId := r.Header.Get("X-Fission-Params-Id")
		params := httpclient.HttpParams{"and": {"OpId^" + opId}, "limit": {"null"}}
		url := fmt.Sprintf("%s/api/op-properties/find", GetDatastoreUrl())

		code, data, err := httpclient.GetR(url, params, nil)
		logger.Debugf("ListOpProperties => %d, %s", code, data)
		errors.Pie(err)

		properties := types.UnmarshalMaps(data)
		response := map[string]interface{}{}
		for _, p := range properties {
			for k, v := range p {
				response[k] = v
			}
		}
		types.MarshalInto(response, w)
	}
}
