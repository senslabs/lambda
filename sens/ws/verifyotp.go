package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/jwt"
	"github.com/senslabs/alpha/sens/logger"
)

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
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.VerifyOtp")
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
		httpclient.WriteError(w, http.StatusBadRequest, errors.New("Category is not provided"))
	}
}
