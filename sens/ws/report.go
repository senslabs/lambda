package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/senslabs/alpha/sens/errors"
	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
	"github.com/senslabs/alpha/sens/types"
)

func GetUserReports(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.GetUserReports")
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		t := r.URL.Query().Get("reportType")
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		params := GetUserParams(w, r, sub["OrgId"].(string))
		if t != "" {
			params["ReportType"] = []string{t}
		}

		if from != "" || to != "" {
			params["span"] = []string{"ReportDate^" + from + "^" + to}
		}
		url := fmt.Sprintf("%s/api/report-views/find", GetDatastoreUrl())
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

func RequestUserReports(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.RequestUserReports")
	if _, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		b, err := ioutil.ReadAll(r.Body)
		errors.Pie(err)
		defer r.Body.Close()
		m := types.UnmarshalMap(b)
		reportType, ok := m["ReportType"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "ReportType is missing")
		}
		url := fmt.Sprintf("%s/api/reports/create", GetDatastoreUrl())
		code, data, err := httpclient.PostR(url, nil, nil, b)
		logger.Debugf("%d, %s", code, data)
		if err != nil {
			logger.Error(err)
			httpclient.WriteError(w, code, err)
		} else {
			userId := r.Header.Get("X-Fission-Params-Id")
			body := fmt.Sprintf(`{
				"subject": "Dozee Update",
				"sender": "narayan@dozee.io",
				"to": [
				  "ankita@dozee.io",
				  "aashit@dozee.io"
				],
				"senderName": "Dozee Support",
				"EmailTemplate": "default",
				"body": {
				  "Data": [
					"The user with user id [%s] requested a report of type [%s]"
				  ]
				}
			  }`, userId, reportType)
			logger.Debug(body)
			code, data, err := httpclient.PostR("http://206.189.141.215:8095/api/v1/email/send", nil, nil, []byte(body))
			logger.Debug(code, data)
			if err != nil {
				logger.Error(err)
			}
			fmt.Fprintf(w, "%s", data)
		}
	}
}
