package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type INIT struct {
	Type   string `json:"type"`
	Secret string `json:"secret"`
}

func main() {
	var err error

	err = godotenv.Load("../.env")
	if err != nil {
		fmt.Println("error when open env")
	}

	http.HandleFunc("/", handle)
	err = http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	var err error
	secret := os.Getenv("VK_API_SECRET")

	_ = r.ParseForm()
	req := r.Body
	_ = req.Close()

	var reqInit INIT
	body, err := ioutil.ReadAll(req)
	b := string(body)

	if err != nil {
		fmt.Println("error on handle request", err.Error())
		return
	}

	_ = json.Unmarshal(body, &reqInit)

	if reqInit.Secret != secret {
		return
	}

	var resBody []byte
	switch reqInit.Type {
	case "confirmation":
		resBody, err = confirmation(b)
		break
	case "message_new":
		resBody, err = messageNew(b)
		break
	}

	if err != nil {
		fmt.Println("error on handle request", err.Error())
		return
	}

	response := http.Response{"200 OK", 200, "HTTP/1.0"}
	response.Body = resBody
	r.Response(resBody)
	fmt.Println(resBody)
}

func confirmation(req string) ([]byte, error) {

	return []byte{}, nil
}

func messageNew(req string) ([]byte, error) {

	return []byte{}, nil
}
