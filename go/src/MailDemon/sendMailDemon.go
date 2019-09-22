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
var err error

const SendQ = "queue_send"
const ApiMessage = "https://api.vk.com/method/messages.send?"
const ApiSave = "https://api.vk.com/method/docs.save?"
const ApiUploadServer = "https://api.vk.com/method/docs.getUploadServer?type=audio_message&peer_id="

func main() {
	err = godotenv.Load(".env")
	if err != nil {
		fmt.Println("error when open env")
		return
	}

	accessToken = os.Getenv("VK_TOKEN")

	isRedis := redisConnect()

	if isRedis {
		handle()
	}
}

func handle() {
	for {
		queueLen := queue.LLen(SendQ).Val()
		if queueLen > 0 && consumerCount < 2 {
			task := queue.LPop(SendQ)
			go consumer(task)
			consumerCount++
		}

		time.Sleep(1 * time.Second)
	}
}

func consumer(task *redis.StringCmd) {
	var queueBody QueueBody
	var taskBody string

	taskBody, err = task.Result()
	if err != nil {
		fmt.Println("error in get task body")
		return
	}

	defer closeConsumer(taskBody)

	err = json.Unmarshal([]byte(taskBody), &queueBody)
	if err != nil {
		fmt.Println("error in decode json")
		return
	}

	message := queueBody.Message
	userId := queueBody.UserId

	var server string
	server, err = getVkUploadServer(userId)
	if err != nil {
		fmt.Println("error get upload server")
		return
	}

	var ownerId, fileId string
	fileId, ownerId, err = uploadVkAudio(server, message)
	if err != nil {
		fmt.Println("error in upload audio")
		return
	}

	err = sendMessage(fileId, ownerId, userId)
}

func sendMessage(fileId, ownerId, userId string) error {
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
		return err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error in json decode")
		return err
	}

	var resMessage messageApi

	err = json.Unmarshal(resBody, resMessage)
	if err != nil {
		fmt.Println("error in json decode")
		return err
	}

	fmt.Println("Сообщение отправлено")

	return nil
}

func uploadVkAudio(server, trackname string) (string, string, error) {
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

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	var saveRes saveResponseApi

	err = json.Unmarshal(body, &saveRes)
	if err != nil {
		fmt.Println("error on decode json saveResponse")
		return "", "", err
	}

	return saveRes.Id, saveRes.OwnerId, nil
}

func getVkUploadServer(peerId string) (string, error) {
	var server uploadApi
	baseUrl := ApiUploadServer
	ver := os.Getenv("VK_API_VERSION")
	Url := baseUrl + peerId + "&v=" + ver + "&access_token=" + accessToken

	resp, err := http.Get(Url)
	if err != nil {
		fmt.Println("error in request vk")
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error in read response vk")
		return "", err
	}

	err = json.Unmarshal(body, &server)
	if err != nil {
		return "", err
	}

	return server.UploadUrl, nil
}

func closeConsumer(task string) {
	if err != nil {
		queue.LPush(SendQ, task)
	}
	consumerCount--
}

func redisConnect() bool {
	redisAddr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	queue = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	res, err := queue.Ping().Result()

	if err != nil || res == "" {
		fmt.Println(`can't connect to redis`)
		return false
	}

	fmt.Println("connected to redis")

	return true
}
