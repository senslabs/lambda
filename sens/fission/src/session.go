package main

import "net/http"

func GetUserSleep(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetUserSleep"))
}

func ListUserSleeps(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ListUserSleeps"))
}
