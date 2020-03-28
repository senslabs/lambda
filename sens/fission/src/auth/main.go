package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/senslabs/alpha/sens/errors"
	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
	"github.com/senslabs/alpha/sens/types"
	"github.com/senslabs/lambda/sens/fission/config"
	"github.com/senslabs/lambda/sens/fission/request"
	"github.com/senslabs/lambda/sens/fission/response"
)

type AuthRequestBody struct {
	Id     string
	Medium string
}

func RequestOtp(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.RequestOtp")
	var reqBody AuthRequestBody
	if err := types.JsonUnmarshalFromReader(r.Body, &reqBody); err != nil {
		logger.Error(err)
		response.WriteError(w, http.StatusInternalServerError, err)
	} else if sessionId, err := requestOtp(reqBody); err != nil {
		logger.Error(err)
		response.WriteError(w, http.StatusInternalServerError, err)
	} else {
		logger.Debug("OTP Sent")
		fmt.Fprintln(w, sessionId)
	}
}

type VerifyRequestBody struct {
	Id               string
	Medium           string
	Session          string
	ConfirmationCode string
}

func VerifyOtp(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.VerifyOtp")
	var reqBody VerifyRequestBody
	if err := types.JsonUnmarshalFromReader(r.Body, &reqBody); err != nil {
		response.WriteError(w, http.StatusInternalServerError, err)
	} else if code, err := verifyOtp(reqBody); err != nil || code != http.StatusOK {
		logger.Error("Code", code, err)
		response.WriteError(w, code, err)
	} else if auth, err := BuildAuth(r, reqBody.Id); err != nil {
		logger.Error(err)
		response.WriteError(w, http.StatusInternalServerError, err)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", auth)
	}
}

func CreateAuth(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.CreateAuth")
	url := fmt.Sprintf("%s/api/auths/create", config.GetDatastoreUrl())
	code, data, err := httpclient.PostR(url, nil, nil, r.Body)
	logger.Debug(code, data)
	if err != nil {
		logger.Error(err)
		response.WriteError(w, code, err)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", data)
	}
}

//Create a user of any category
func createUser(w http.ResponseWriter, r *http.Request, category string) error {
	url := fmt.Sprintf("%s/api/%ss/create", config.GetDatastoreUrl(), strings.ToLower(category))
	code, data, err := httpclient.PostR(url, nil, nil, r.Body)
	logger.Debug(code, data)
	if err != nil {
		return httpclient.WriteError(w, code, err)
	}
	fmt.Fprintf(w, "%s", data)
	return nil
}

//Map that category user to auth
// func mapUserAuth(w http.ResponseWriter, r *http.Request, category string, categoryId string) error {
// 	url := fmt.Sprintf("%s/api/%s-auths/create", config.GetDatastoreUrl(), strings.ToLower(category))
// 	authId := r.Header.Get("x-sens-auth-id")
// 	body := fmt.Sprintf(`{"%sId": "%s", "AuthId":"%s"}`, category, categoryId, authId)
// 	code, data, err := httpclient.PostR(url, nil, nil, body)
// 	logger.Debug(code, data)
// 	if err != nil {
// 		logger.Error(err)
// 		logger.Error("Failed in Mapping", category, "AuthId:", authId, category+"Id:", authId)
// 	}
// 	return err
// }

//Get the category user details
func getUserDetail(w http.ResponseWriter, r *http.Request, category string) {
	id := request.GetPathParam(r, "id")
	url := fmt.Sprintf("%s/api/%s-detail-views/%s/get", config.GetDatastoreUrl(), strings.ToLower(category), id)
	code, data, err := httpclient.GetR(url, nil, nil)
	logger.Debugf("Code: %d, Data: %s", code, data)
	if err != nil {
		logger.Error(err)
		response.WriteError(w, code, err)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", data)
	}
}

func GetOrgDetail(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.GetOrgDetail")
	getUserDetail(w, r, "Org")
}

func GetOpDetail(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.GetOpDetail")
	getUserDetail(w, r, "Op")
}

func GetUserDetail(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.GetUserDetail")
	getUserDetail(w, r, "User")
}

func CreateOrg(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.CreateOrg")
	createUser(w, r, "Org")
}

func CreateOp(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.CreateOp")
	createUser(w, r, "Op")
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.CreateUser")
	createUser(w, r, "User")
}

type TwilioSendOtpResponse struct {
	ServiceSid string `json:"service_sid"`
}

func requestOtp(reqBody AuthRequestBody) (string, error) {
	if reqBody.Medium == "Mobile" {
		body := fmt.Sprintf("To=%s&Channel=sms&Locale=en", url.QueryEscape(reqBody.Id))
		url := "https://verify.twilio.com/v2/Services/VAd156b7c4b609261239603a320c3af4e2/Verifications"

		var twilioResponse TwilioSendOtpResponse
		code, err := httpclient.Post(url, nil, httpclient.HttpParams{
			"Authorization": {"Basic QUM2MmFlOWU0N2I2MTI2M2YyZDQwYzdjYjhjMzMyNzU4OTo4MTg4MGNhMTBmMjMxMGUxNjdlZGI1YTRmZGVjMDUxMg=="},
			"Content-Type":  {"application/x-www-form-urlencoded"},
		}, []byte(body), &twilioResponse)
		logger.Debugf("%d, %v", code, twilioResponse)
		if err != nil || code != http.StatusCreated {
			logger.Error("HTTP response code:", code, err)
			return "", err
		} else {
			return twilioResponse.ServiceSid, nil
		}
	}
	return "", errors.New(http.StatusInternalServerError, "Only Mobiles are supported for sending OTP")
}

type TwilioVerifyOtpResponse struct {
	Status string `json:"status"`
}

func verifyOtp(reqBody VerifyRequestBody) (int, error) {
	body := fmt.Sprintf("To=%s&Code=%s", url.QueryEscape(reqBody.Id), reqBody.ConfirmationCode)
	url := fmt.Sprintf("https://verify.twilio.com/v2/Services/%s/VerificationCheck", reqBody.Session)
	var twilioResponse TwilioVerifyOtpResponse
	code, data, err := httpclient.PostR(url, nil, httpclient.HttpParams{
		"Authorization": {"Basic QUM2MmFlOWU0N2I2MTI2M2YyZDQwYzdjYjhjMzMyNzU4OTo4MTg4MGNhMTBmMjMxMGUxNjdlZGI1YTRmZGVjMDUxMg=="},
		"Content-Type":  {"application/x-www-form-urlencoded"},
	}, []byte(body))
	if err != nil {
		logger.Error(err)
		return code, err
	}
	if code == http.StatusOK {
		err = json.Unmarshal(data, &twilioResponse)
	}
	if err != nil || code != http.StatusOK || twilioResponse.Status != "approved" {
		logger.Error("Twilio confrmation code entered does not seem to be correct")
		return code, errors.New(0, "Twilio confrmation code entered does not seem to be correct")
	}
	return code, nil
}

func buildAuth(r *http.Request, auths []byte, id string) ([]byte, error) {
	var subs []map[string]interface{}
	if err := json.Unmarshal(auths, &subs); err != nil {
		logger.Error(err)
		return nil, errors.FromError(errors.GO_ERROR, err)
	}
	sub := subs[0]
	logger.Debugf("%#v", sub)
	category := request.GetQueryParam(r, "category")
	// categoryIdField := strings.Title(category) + "Id"
	sub["Category"] = category
	// url := fmt.Sprintf("%s/api/%ss/find", config.GetDatastoreUrl(), sub["Category"])
	// params := httpclient.HttpParams{"and": {"AuthId^" + sub["Id"]}, "limit": {1}}
	// httpclient.GetR(url, params, nil)
	return json.Marshal(sub)
}

func BuildAuth(r *http.Request, id string) ([]byte, error) {
	or := httpclient.HttpParams{"or": {"Mobile^" + id, "Email^" + id, "Social^" + id}, "limit": {"1"}}
	logger.Debugf("OR: %#v", or)
	url := fmt.Sprintf("%s/api/auths/find", config.GetDatastoreUrl())
	code, auths, err := httpclient.GetR(url, or, nil)
	logger.Debugf("%d, %s", code, auths)
	if err != nil || code != http.StatusOK {
		logger.Error("HTTP response code:", code, err)
		return nil, err
	} else if bytes.Equal(auths, []byte("[]")) {
		return []byte(fmt.Sprintf(`[{"Id": "%s", "Exists": false}]`, id)), nil
	} else {
		return buildAuth(r, auths, id)
	}
}

func main() {}
