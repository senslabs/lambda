package response

import "net/http"

func WriteError(w http.ResponseWriter, code int, err error) {
	http.Error(w, err.Error(), code)
}
