package config

import "os"

func GetDatastoreUrl() string {
	baseUrl := os.Getenv("DATASTORE_BASE_URL")
	if baseUrl == "" {
		return "http://datastore.senslabs.me"
	}
	return baseUrl
}