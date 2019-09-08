package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

type QueueBody struct {
	UserId  string `json:"user_id"`
	Message string `json:"message"`
}

type uploadApi struct {
	UploadUrl string `json:"upload_url"`
}

type saveApi struct {
	File string `json:"file"`
}

type saveResponseApi struct {
	Id      string `json:"id"`
	OwnerId string `json:"owner_id"`
}

type messageApi struct {
	Response string `json:"response"`
}

var queue *redis.Client
var consumerCount uint8 = 0
var mu sync.Mutex
var accessToken string

const SendQ = "queue_send"
const ApiMessage = "https://api.vk.com/method/messages.send?"
const ApiSave = "https://api.vk.com/method/docs.save?"
const ApiUploadServer = "https://api.vk.com/method/docs.getUploadServer?type=audio_message&peed_id="

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("error when open env")
		return
	}

	accessToken = os.Getenv("VK_TOKEN")

	isRedis := redisConnect()

	if isRedis {
		handle()
	} else {
		fmt.Println(`can't connect to redis`)
	}
}

func handle() {
	for {
		if consumerCount < 2 {
			consumerCount++
			go consumer()
		}

		time.Sleep(1)
	}
}

func consumer() {
	defer closeConsumer()
	var queueBody QueueBody

	task := queue.Get(SendQ)
	taskBody, err := task.Result()
	if err != nil {
		fmt.Println("error in get task body")
		return
	}

	err = json.Unmarshal([]byte(taskBody), &queueBody)

	if err != nil {
		fmt.Println("error in decode json")
		return
	}

	body := queueBody.Message
	userId := queueBody.UserId

	server := getVkUploadServer(userId)

	fileId, ownerId, err := uploadVkAudio(server, body)
	if err != nil {
		fmt.Println("error in upload audio")
		return
	}

	sendMessage(fileId, ownerId, userId)
}

func sendMessage(fileId, ownerId, userId string) {
	randomId := string(rand.Int() + rand.Intn(1000))
	document := fmt.Sprintf("doc%s_%s", fileId, ownerId)

	urlArgs := url.Values{}
	urlArgs.Add("user_id", userId)
	urlArgs.Add("random_id", randomId)
	urlArgs.Add("attachment", document)
	urlArgs.Add("access_token", accessToken)
	urlPart := urlArgs.Encode()

	fullurl := ApiMessage + urlPart

	res, err := http.Get(fullurl)
	if err != nil {
		fmt.Println("error in send message")
		return
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error in json decode")
	}

	var resMessage messageApi

	err = json.Unmarshal(resBody, resMessage)
	if err != nil {
		fmt.Println("error in json decode")
	}

	return
}

func uploadVkAudio(server, trackname string) (string, string, error) {
	var err error
	fileDir, _ := os.Getwd()
	filePath := path.Join(fileDir, trackname)

	file, _ := os.Open(filePath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		fmt.Println("error create form file")
		return "", "", err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println("error file close")
		return "", "", err
	}

	err = writer.Close()
	if err != nil {
		fmt.Println("error file close")
		return "", "", err
	}

	r, err := http.NewRequest("POST", server, body)
	if err != nil {
		fmt.Println("error create request")
		return "", "", err
	}

	r.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}

	res, err := client.Do(r)
	if err != nil {
		fmt.Println("error send request")
		return "", "", err
	}
	var uploadFile saveApi

	resBody, err := ioutil.ReadAll(res.Body)

	err = json.Unmarshal([]byte(resBody), &uploadFile)

	fileId, ownerId, err := saveFileVk(uploadFile.File)
	if err != nil {
		fmt.Println("error on save file in vk")
		ownerId = ""
		fileId = ""
	}

	err = os.Remove(filePath)
	if err != nil {
		fmt.Println("error remove file")
	}

	err = file.Close()
	if err != nil {
		fmt.Println("error file close")
	}

	return fileId, ownerId, err
}

func saveFileVk(file string) (string, string, error) {
	fullUrl := ApiSave + "&file=" + file + "&access_token=" + accessToken
	res, err := http.Get(fullUrl)

	if err != nil {
		return "", "", err
	}
	body, _ := ioutil.ReadAll(res.Body)

	var saveRes saveResponseApi

	err = json.Unmarshal(body, &saveRes)
	if err != nil {
		fmt.Println("error on decode json saveResponse")
		return "", "", err
	}

	return saveRes.Id, saveRes.OwnerId, nil
}

func getVkUploadServer(peerId string) string {
	var server uploadApi
	baseUrl := ApiUploadServer
	ver := os.Getenv("VK_API_VERSION")
	Url := baseUrl + peerId + "&v=" + ver + "&access_token=" + accessToken

	resp, err := http.Get(Url)

	if err != nil {
		fmt.Println("error in request vk")
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("error in read response vk")
		return ""
	}

	err = json.Unmarshal(body, &server)

	return server.UploadUrl
}

func closeConsumer() {
	mu.Lock()
	consumerCount--
	mu.Unlock()
}

func redisConnect() bool {
	queue = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := queue.Ping().Result()

	if err != nil {
		return false
	}

	return true
}
