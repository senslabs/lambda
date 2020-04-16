package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
)

type AuthRequestBody struct {
	Medium      string
	MediumValue string
}

type TwilioSendOtpResponse struct {
	ServiceSid string `json:"service_sid"`
}

func RequestOtp(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.RequestOtp")
	var body string
	var reqBody AuthRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else if reqBody.Medium == "Mobile" {
		body = fmt.Sprintf("To=%s&Channel=sms&Locale=en", url.QueryEscape(reqBody.MediumValue))
	} else if reqBody.Medium == "Email" {
		body = fmt.Sprintf("To=%s&Channel=email&Locale=en", url.QueryEscape(reqBody.MediumValue))
	}
	if body != "" {
		url := "https://verify.twilio.com/v2/Services/VAd156b7c4b609261239603a320c3af4e2/Verifications"
		var twilioResponse TwilioSendOtpResponse
		logger.Debugf("%s", body)
		code, err := httpclient.Post(url, nil, httpclient.HttpParams{
			"Authorization": {"Basic QUM2MmFlOWU0N2I2MTI2M2YyZDQwYzdjYjhjMzMyNzU4OTo4MTg4MGNhMTBmMjMxMGUxNjdlZGI1YTRmZGVjMDUxMg=="},
			"Content-Type":  {"application/x-www-form-urlencoded"},
		}, []byte(body), &twilioResponse)
		logger.Debugf("%d, %v", code, twilioResponse)
		if err != nil || code != http.StatusCreated {
			logger.Error("HTTP response code:", code, err)
			httpclient.WriteError(w, http.StatusInternalServerError, err)
		} else {
			logger.Debug("OTP Sent")
			fmt.Fprintf(w, `{"SessionId":"%s"}`, twilioResponse.ServiceSid)
		}
	} else {
		logger.Error("Didn't receive the Medium")
		httpclient.WriteError(w, http.StatusInternalServerError, errors.New("Didn't receive the Medium"))
	}
}

func VerifyAuth(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.VerifyAuth")
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
