package gcloudfn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/jwt"
	"github.com/senslabs/alpha/sens/logger"
)

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.RefreshToken")
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		authorization = r.Header.Get("authorization")
	}
	token := strings.TrimPrefix(authorization, "Bearer ")
	if authorization == "" {
		w.WriteHeader(http.StatusUnauthorized)
	} else if token, err := jwt.RefreshAccessToken(token); err != nil {
		logger.Error(err)
		httpclient.WriteError(w, http.StatusInternalServerError, err)
	} else {
		fmt.Fprintf(w, `{"AccessToken":"%s"}`, token)
	}
}

func getAuthSubject(r *http.Request) (map[string]interface{}, error) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		authorization = r.Header.Get("authorization")
	}
	accessToken := strings.TrimPrefix(authorization, "Bearer ")
	if accessToken == "" {
		logger.Debug("Access Token Empty")
		//Specific handling to understand that if the client has sent Acess Token or Not
		return nil, nil
	}
	if sub, err := jwt.VerifyToken(accessToken); err != nil {
		logger.Error(err)
		return nil, err
	} else {
		logger.Debugf("%#v", sub)
		return sub, nil
	}
}

const (
	OPEN = iota
	KEY
	SENS
)

func IsSens(w http.ResponseWriter, r *http.Request) (bool, error) {
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		return false, err
	} else if sub != nil && sub["IsSens"] != nil {
		logger.Debugf("%#v", sub)
		return sub["IsSens"].(bool), nil
	} else {
		return false, err
	}
}

type org struct {
	OrgName string
	OrgId   string
}

func ChangeRoleToOp(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ChangeRoleToOp")
	var o org
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		logger.Error(err)
		httpclient.WriteInternalServerError(w, err)
	} else {
		sub["OrgId"] = o.OrgId
		sub["OrgName"] = o.OrgName
		generateLoginTokens(w, r, sub, "auth")
	}
}

func ChangeRoleToAuth(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	logger.InitLogger("wsproxy.ChangeRoleToAuth")
	if sub, err := getAuthSubject(r); err != nil {
		logger.Error(err)
		httpclient.WriteUnauthorizedError(w, err)
	} else {
		delete(sub, "OrgId")
		delete(sub, "OrgName")
		generateLoginTokens(w, r, sub, "auth")
	}
}
