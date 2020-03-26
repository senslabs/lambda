package main

import (
	"fmt"
	"net/http"

	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
	"github.com/senslabs/lambda/sens/fission/config"
	"github.com/senslabs/lambda/sens/fission/request"
	"github.com/senslabs/lambda/sens/fission/response"
)

func ListOpDevices(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.ListOpDevices")
	opId := request.GetSensHeaderValue(r, "Op-id")
	and := httpclient.HttpParams{"and": {"OrgId^" + orgId, "Status^REGISTERED"}, "limit": {"100"}}
	url := fmt.Sprintf("%s/api/devices/find", config.GetDatastoreUrl())
	code, data, err := httpclient.GetR(url, and, nil)
	logger.Debug(code, data)
	if err != nil {
		logger.Error(err)
		response.WriteError(w, code, err)
	} else {
		fmt.Fprintf(w, "%s", data)
	}
}

func ListOrgUsers(w http.ResponseWriter, r *http.Request) {
	logger.InitLogger("sens.lambda.ListOrgUsers")
	orgId := request.GetHeaderValue(r, "x-sens-org-id")
	and := httpclient.HttpParams{"and": {"OrgId^" + orgId}, "limit": {"100"}}
	url := fmt.Sprintf("%s/api/users/find", config.GetDatastoreUrl())
	code, data, err := httpclient.GetR(url, and, nil)
	logger.Debug(code, data)
	if err != nil {
		logger.Error(err)
		response.WriteError(w, code, err)
	} else {
		fmt.Fprintf(w, "%s", data)
	}
}

func main() {}
