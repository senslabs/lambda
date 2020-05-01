package httpclient

import (
	"io/ioutil"
	"net/http"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/senslabs/alpha/sens/errors"
	"github.com/senslabs/alpha/sens/logger"
	"github.com/senslabs/alpha/sens/types"
)

type HttpParams map[string][]string

func prepare(req *retryablehttp.Request, params HttpParams, headers HttpParams) {
	query := req.URL.Query()
	for k, v := range params {
		for _, v := range v {
			query.Add(k, v)
		}
	}
	req.URL.RawQuery = query.Encode()
	for k, v := range headers {
		for _, v := range v {
			req.Header.Add(k, v)
		}
	}
}

func PerformR(req *retryablehttp.Request, params HttpParams, headers HttpParams) (int, []byte, error) {
	prepare(req, params, headers)
	client := retryablehttp.NewClient()
	if res, err := client.Do(req); err != nil {
		return http.StatusInternalServerError, nil, errors.FromError(errors.GO_ERROR, err)
	} else if b, err := ioutil.ReadAll(res.Body); err != nil {
		return http.StatusInternalServerError, nil, errors.FromError(errors.GO_ERROR, err)
	} else {
		defer res.Body.Close()
		return res.StatusCode, b, nil
	}
}

func GetR(url string, params HttpParams, headers HttpParams) (int, []byte, error) {
	if req, err := retryablehttp.NewRequest("GET", url, nil); err != nil {
		logger.Error(err)
		return http.StatusInternalServerError, nil, errors.FromError(errors.GO_ERROR, err)
	} else {
		return PerformR(req, params, headers)
	}
}

func PostR(url string, params HttpParams, headers HttpParams, rawBody interface{}) (int, []byte, error) {
	if req, err := retryablehttp.NewRequest("POST", url, rawBody); err != nil {
		logger.Error(err)
		return http.StatusInternalServerError, nil, errors.FromError(errors.GO_ERROR, err)
	} else {
		return PerformR(req, params, headers)
	}
}

func Perform(req *retryablehttp.Request, params HttpParams, headers HttpParams, response interface{}) (int, error) {
	prepare(req, params, headers)
	client := retryablehttp.NewClient()
	if res, err := client.Do(req); err != nil {
		return http.StatusInternalServerError, errors.FromError(errors.GO_ERROR, err)
	} else if err := types.JsonUnmarshalFromReader(res.Body, response); err != nil {
		return http.StatusInternalServerError, err
	} else {
		defer res.Body.Close()
		return res.StatusCode, nil
	}
}

func Get(url string, params HttpParams, headers HttpParams, response interface{}) (int, error) {
	if req, err := retryablehttp.NewRequest("GET", url, nil); err != nil {
		logger.Error(err)
		return http.StatusInternalServerError, errors.FromError(errors.GO_ERROR, err)
	} else {
		return Perform(req, params, headers, response)
	}
}

func Post(url string, params HttpParams, headers HttpParams, rawBody interface{}, response interface{}) (int, error) {
	if req, err := retryablehttp.NewRequest("POST", url, rawBody); err != nil {
		logger.Error(err)
		return http.StatusInternalServerError, errors.FromError(errors.GO_ERROR, err)
	} else {
		return Perform(req, params, headers, response)
	}
}

func WriteError(w http.ResponseWriter, code int, err error) error {
	if err == nil {
		http.Error(w, "nil", code)
	} else {
		http.Error(w, err.Error(), code)
	}
	return err
}

func WriteInternalServerError(w http.ResponseWriter, err error) error {
	return WriteError(w, http.StatusInternalServerError, err)
}

func WriteUnauthorizedError(w http.ResponseWriter, err error) error {
	return WriteError(w, http.StatusUnauthorized, err)
}
