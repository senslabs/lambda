package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/senslabs/alpha/sens/errors"
	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
	"github.com/senslabs/alpha/sens/types"
	"github.com/senslabs/lambda/sens/fission/response"
)

type AuthRequestBody struct {
	Id      string
	Medium  string
	Otp     string
	Session string
}

const DATASTORE_BASE_URL = "http://datastore.zonea.senslabs.io:9804"

func RequestOtp(w http.ResponseWriter, r *http.Request) {
	var reqBody AuthRequestBody
	if err := types.JsonUnmarshalFromReader(r.Body, &reqBody); err != nil {
		logger.Error(err)
		response.WriteError(w, http.StatusInternalServerError, err)
	} else if sessionId, err := requestOtp(reqBody); err != nil {
		logger.Error(err)
		response.WriteError(w, http.StatusInternalServerError, err)
	} else {
		logger.Debug("OTP Sent")
		fmt.Println(w, sessionId)
	}
}

type VerifyRequestBody struct {
	Id               string
	Medium           string
	Otp              string
	Session          string
	ConfirmationCode string
}

func VerifyOtp(w http.ResponseWriter, r *http.Request) {
	var reqBody VerifyRequestBody
	if err := types.JsonUnmarshalFromReader(r.Body, &reqBody); err != nil {
		logger.Error(err)
		response.WriteError(w, http.StatusInternalServerError, err)
	} else if verified, err := verifyOtp(reqBody); err != nil || !verified {
		logger.Error("Verified:", verified, err)
		response.WriteError(w, http.StatusInternalServerError, err)
	} else if auth, err := LoadAuth(reqBody.Id); err != nil {
		logger.Error(err)
		response.WriteError(w, http.StatusInternalServerError, err)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, auth)
	}
}

func CreateAuth(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%s/api/auths/create", DATASTORE_BASE_URL)
	code, data, err := httpclient.PostR(url, nil, nil, r.Body)
	logger.Debug(code, data)
	if err != nil {
		logger.Error(err)
		response.WriteError(w, code, err)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, data)
	}
}

//Create a user of any category
func createUser(w http.ResponseWriter, r *http.Request, category string) {
	url := fmt.Sprintf("%s/api/%ss/create", DATASTORE_BASE_URL, strings.ToLower(category))
	code, data, err := httpclient.PostR(url, nil, nil, r.Body)
	logger.Debug(code, data)
	if err != nil {
		logger.Error(err)
		response.WriteError(w, code, err)
	} else if err := mapUserAuth(w, r, category, string(data)); err != nil {
		logger.Error(err)
		response.WriteError(w, code, err)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, data)
	}
}

//Map that category user to auth
func mapUserAuth(w http.ResponseWriter, r *http.Request, category string, categoryId string) error {
	url := fmt.Sprintf("%s/api/%s-auths/create", DATASTORE_BASE_URL, strings.ToLower(category))
	authId := r.Header.Get("x-sens-auth-id")
	body := fmt.Sprintf(`{"%sId": "%s", "AuthId":"%s"}`, category, categoryId, authId)
	code, data, err := httpclient.PostR(url, nil, nil, body)
	logger.Debug(code, data)
	if err != nil {
		logger.Error(err)
		logger.Error("Failed in Mapping", category, "AuthId:", authId, category+"Id:", authId)
	}
	return err
}

func CreateOrg(w http.ResponseWriter, r *http.Request) {
	createUser(w, r, "Org")
}

func CreateOp(w http.ResponseWriter, r *http.Request) {
	createUser(w, r, "Op")
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	createUser(w, r, "User")
}

type TwilioSendOtpResponse struct {
	SessionSid string
}

func requestOtp(reqBody AuthRequestBody) (string, error) {
	if reqBody.Medium == "Mobile" {
		body := fmt.Sprintf("To=%s&Channel=sms&Locale=en", url.QueryEscape(reqBody.Id))
		url := "https://verify.twilio.com/v2/Services/VAd156b7c4b609261239603a320c3af4e2/Verifications"

		var twilioResponse TwilioSendOtpResponse
		code, err := httpclient.Post(url, nil, map[string]string{
			"Authorization": "Basic QUM2MmFlOWU0N2I2MTI2M2YyZDQwYzdjYjhjMzMyNzU4OTo4MTg4MGNhMTBmMjMxMGUxNjdlZGI1YTRmZGVjMDUxMg==",
			"Content-Type":  "application/x-www-form-urlencoded",
		}, []byte(body), &twilioResponse)
		logger.Debugf("%d, %v", code, twilioResponse)
		if err != nil || code != http.StatusOK {
			logger.Error("HTTP response code:", code, err)
			return "", err
		} else {
			return twilioResponse.SessionSid, nil
		}
	}
	return "", errors.New(http.StatusInternalServerError, "Only Mobiles are supported for sending OTP")
}

type TwilioVerifyOtpResponse struct {
	Status string `json:"status"`
}

func verifyOtp(reqBody VerifyRequestBody) (bool, error) {
	body := fmt.Sprintf("To=%s&Code=%s", url.QueryEscape(reqBody.Id), reqBody.ConfirmationCode)
	url := fmt.Sprintf("https://verify.twilio.com/v2/Services/%s/VerificationCheck", reqBody.Session)
	var twilioResponse TwilioVerifyOtpResponse
	code, err := httpclient.Post(url, nil, map[string]string{
		"Authorization": "Basic QUM2MmFlOWU0N2I2MTI2M2YyZDQwYzdjYjhjMzMyNzU4OTo4MTg4MGNhMTBmMjMxMGUxNjdlZGI1YTRmZGVjMDUxMg==",
		"Content-Type":  "application/x-www-form-urlencoded",
	}, []byte(body), &twilioResponse)
	logger.Debugf("%d, %v", code, twilioResponse)
	if err != nil {
		logger.Error(err)
		return false, err
	}
	if twilioResponse.Status != "approved" {
		logger.Error("Twilio confrmation code entered does not seem to be correct")
		return false, errors.New(0, "Twilio confrmation code entered does not seem to be correct")
	}
	return true, nil
}

func LoadAuth(id string) ([]byte, error) {
	or := url.Values{"Mobile": []string{id}, "Email": []string{id}, "Social": []string{id}}
	url := fmt.Sprintf("%s/api/auths/find", DATASTORE_BASE_URL)
	code, auths, err := httpclient.GetR(url, or, nil)
	logger.Debugf("%d, %s", code, auths)
	if err != nil || code != http.StatusOK {
		logger.Error("HTTP response code:", code, err)
		return nil, err
	} else if bytes.Equal(auths, []byte("[]")) {
		if authId, err := uuid.NewRandom(); err != nil {
			logger.Error(err)
			return nil, errors.FromError(errors.GO_ERROR, err)
		} else {
			return []byte(fmt.Sprintf(`[{"Id": %s, "Exists", false}]`, authId.String())), nil
		}
	} else {
		return auths, nil
	}
}
