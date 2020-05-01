package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
)

//gcloud functions deploy NAME --source https://source.developers.google.com/projects/PROJECT_ID/repos/REPOSITORY_ID/moveable-aliases/master/paths/SOURCE --runtime RUNTIME TRIGGER... [FLAGS...]
//This is same a ListAlerts
func ListUserAlerts(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListUserAlerts")
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
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func ListLatestAlerts(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListLatestAlerts")
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		params := GetUserParams(w, r, sub["OrgId"].(string))
		url := fmt.Sprintf("%s/api/org-latest-alert-views/find", GetDatastoreUrl())
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

func CreateAlertRule(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.CreateAlertRule")
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		url := fmt.Sprintf("%s/api/alert-rules/create", GetDatastoreUrl())
		code, data, err := httpclient.PostR(url, nil, nil, r.Body)
		defer r.Body.Close()
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func CreateAlertEscalation(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.CreateAlertEscalations")
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		url := fmt.Sprintf("%s/api/alert-escalations/create", GetDatastoreUrl())
		code, data, err := httpclient.PostR(url, nil, nil, r.Body)
		defer r.Body.Close()
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func UpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.UpdateAlertRule")
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		id := r.Header.Get("X-Fission-Params-Id")
		url := fmt.Sprintf("%s/api/alert-rules/%s/update", GetDatastoreUrl(), id)
		code, data, err := httpclient.PostR(url, nil, nil, r.Body)
		defer r.Body.Close()
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func UpdateAlertEscalation(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.UpdateAlertEscalation")
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		id := r.Header.Get("X-Fission-Params-Id")
		url := fmt.Sprintf("%s/api/alert-escalations/%s/update", GetDatastoreUrl(), id)
		code, data, err := httpclient.PostR(url, nil, nil, r.Body)
		defer r.Body.Close()
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func UpdateAlert(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.UpdateAlert")
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		id := r.Header.Get("X-Fission-Params-Id")
		url := fmt.Sprintf("%s/api/alerts/%s/update", GetDatastoreUrl(), id)
		code, data, err := httpclient.PostR(url, nil, nil, r.Body)
		defer r.Body.Close()
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func ListAlertRules(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListAlertRules")
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		url := fmt.Sprintf("%s/api/alert-rules/create", GetDatastoreUrl())
		code, data, err := httpclient.PostR(url, nil, nil, r.Body)
		defer r.Body.Close()
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}

func ListAlertEscalations(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ListAlertEscalations")
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		url := fmt.Sprintf("%s/api/alert-escalations/create", GetDatastoreUrl())
		code, data, err := httpclient.PostR(url, nil, nil, r.Body)
		defer r.Body.Close()
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			fmt.Fprintf(w, "%s", data)
		}
	}
}
