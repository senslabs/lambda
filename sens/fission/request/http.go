package request

import (
	"fmt"
	"net/http"
	"net/textproto"
)

func GetPathParam(r *http.Request, key string) string {
	key = textproto.CanonicalMIMEHeaderKey(key)
	key = fmt.Sprintf("%s-%s", "X-Fission-Params", key)
	return r.Header.Get(key)
}

func GetHeaderValue(r *http.Request, key string) string {
	key = textproto.CanonicalMIMEHeaderKey(key)
	key = fmt.Sprintf("%s-%s", "X-Sens", key)
	return r.Header.Get(key)
}

func GetQueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}
