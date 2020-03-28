package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
	"github.com/senslabs/lambda/sens/fission/config"
	"github.com/senslabs/lambda/sens/fission/request"
	"github.com/senslabs/lambda/sens/fission/response"
)

func ListOrgDevices(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.ListOrgDevices")
	orgId := request.GetSensHeaderValue(r, "Org-Id")
	status := request.GetQueryParam(r, "status")
	and := httpclient.HttpParams{"and": {"OrgId^" + orgId, "Status^" + strings.ToUpper(status)}, "limit": {"100"}}
	url := fmt.Sprintf("%s/api/devices/find", config.GetDatastoreUrl())
	code, data, err := httpclient.GetR(url, and, nil)
	logger.Debugf("%d, %#v", code, data)
	if err != nil {
		logger.Error(err)
		response.WriteError(w, code, err)
	} else {
		fmt.Fprintf(w, "%s", data)
	}
}

func ListOrgUsers(w http.ResponseWriter, r *http.Request) {
	os.Setenv("LOG_STORE", "fluentd")
	os.Setenv("FLUENTD_HOST", "fluentd.senslabs.me")
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("sens.lambda.ListOrgUsers")
	orgId := request.GetHeaderValue(r, "x-sens-org-id")
	and := httpclient.HttpParams{"and": {"OrgId^" + orgId}, "limit": {"100"}}
	url := fmt.Sprintf("%s/api/user-detail-views/find", config.GetDatastoreUrl())
	code, data, err := httpclient.GetR(url, and, nil)
	logger.Debugf("ListOrgUsers: %d, %s", code, data)
	if err != nil {
		logger.Error(err)
		response.WriteError(w, code, err)
	} else {
		fmt.Fprintf(w, "%s", data)
	}
}

func main() {}
