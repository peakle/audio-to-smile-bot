package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", handle)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
}
