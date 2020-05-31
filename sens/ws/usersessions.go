package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
	"github.com/senslabs/alpha/sens/types"
)

func splitValues(session []map[string]interface{}) {
	for i, s := range session {
		var timestamps [][]interface{}
		var values [][]interface{}
		timestamp := s["Timestamps"].([]interface{})
		value := s["Values"].([]interface{})
		if len(timestamp) >= 2 {
			prev := 0
			l := len(timestamp)
			for k := 1; k < l; k++ {
				if timestamp[k].(float64)-timestamp[k-1].(float64) > 240 {
					timestamps = append(timestamps, timestamp[prev:k])
					logger.Debugf("Timestamps: %#v", timestamps)
					values = append(values, value[prev:k])
					prev = k
				}
			}
			logger.Debug("Here")
			timestamps = append(timestamps, timestamp[prev:l])
			values = append(values, value[prev:l])
		}
		session[i]["Timestamps"] = timestamps
		session[i]["Values"] = values
	}
}

// api/user/sessions/{id}/get
func GetUserSession(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.GetUserSession")
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		id := r.Header.Get("X-Fission-Params-Id")
		params := httpclient.HttpParams{"and": {"OrgId^" + sub["OrgId"].(string), "SessionId^" + id}, "limit": {"null"}}
		url := fmt.Sprintf("%s/api/org-session-detail-views/find", GetDatastoreUrl())

		var session []map[string]interface{}
		// var properties map[string]string
		code, err := httpclient.Get(url, params, nil, &session)
		splitValues(session)

		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			types.MarshalInto(session, w)
		}
	}
}

// api/user/sessions/{id}/events/get
func GetUserSessionEvents(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.GetUserSession")
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		id := r.Header.Get("X-Fission-Params-Id")
		params := httpclient.HttpParams{"and": {"OrgId^" + sub["OrgId"].(string), "SessionId^" + id}, "limit": {"null"}}
		url := fmt.Sprintf("%s/api/org-session-event-detail-views/find", GetDatastoreUrl())

		code, data, err := httpclient.GetR(url, params, nil)

		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func GetSleepTrend(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.GetDatedSleeps")
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		id := r.Header.Get("X-Fission-Params-Id")
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		params := httpclient.HttpParams{"From": {from}, "To": {to}}
		url := fmt.Sprintf("%s/api/ext/users/%s/trends", GetDatastoreUrl(), id)
		code, data, err := httpclient.GetR(url, params, nil)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func GetUserDatedSession(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.GetUserDateSession")
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		id := r.Header.Get("X-Fission-Params-Id")
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		params := httpclient.HttpParams{"span": {"Date^" + from + "^" + to}, "and": {"UserId^" + id}, "limit": {"null"}}
		url := fmt.Sprintf("%s/api/user-dated-session-views/find", GetDatastoreUrl())
		code, data, err := httpclient.GetR(url, params, nil)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func GetLongestSleepTrend(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.GetLongestSleepTrend")
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		id := r.Header.Get("X-Fission-Params-Id")
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		params := httpclient.HttpParams{"span": {"Date^" + from + "^" + to}, "and": {"UserId^" + id}, "limit": {"null"}}
		url := fmt.Sprintf("%s/api/longest-sleep-trend-views/find", GetDatastoreUrl())
		code, data, err := httpclient.GetR(url, params, nil)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func ListUserBaselines(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListUserBaselines")
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		id := r.Header.Get("X-Fission-Params-Id")
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		params := httpclient.HttpParams{"span": {"CreatedAt^" + from + "^" + to}, "and": {"UserId^" + id}, "limit": {"null"}}
		url := fmt.Sprintf("%s/api/user-baseline-views/find", GetDatastoreUrl())
		code, data, err := httpclient.GetR(url, params, nil)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}
