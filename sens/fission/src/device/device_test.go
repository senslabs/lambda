package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/senslabs/alpha/sens/httpclient"
	"github.com/senslabs/alpha/sens/logger"
)

func TestDevice(t *testing.T) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger.InitLogger("")
	url := fmt.Sprintf("%s%s", "http://localhost:9804", "/api/devices/find")
	and := httpclient.HttpParams{"span": {"CreatedAt^1585032690^1585032791"}, "limit": {"10"}}
	code, data, err := httpclient.GetR(url, and, nil)
	fmt.Println(code, string(data), err)
}
