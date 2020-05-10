package gcloudfn

import "os"

func GetDatastoreUrl() string {
	url := os.Getenv("DATASTORE_BASE_URL")
	if url == "" {
		return "http://datastore.senslabs.me"
	}
	return url
}
