package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
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
var v string
var err error
var logFile *os.File
var logger *log.Logger

const SendQ = "queue_send"
const ApiMessage = "https://api.vk.com/method/messages.send?"
const ApiSave = "https://api.vk.com/method/docs.save?"
const ApiUploadServer = "https://api.vk.com/method/docs.getUploadServer?"

func main() {
	logging()
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("error when open env")
		return
	}

	accessToken = os.Getenv("VK_TOKEN")
	v = os.Getenv("VK_API_VERSION")

	isRedis := redisConnect()
	defer destruct()

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
		logger.Println("error in get task body " + err.Error())
		return
	}

	defer closeConsumer(taskBody)

	err = json.Unmarshal([]byte(taskBody), &queueBody)
	if err != nil {
		logger.Println("error in decode json " + err.Error())
		return
	}

	message := queueBody.Message
	userId := queueBody.UserId

	var server string
	server, err = getVkUploadServer(userId)
	if err != nil {
		return
	}

	var ownerId, fileId string
	fileId, ownerId, err = uploadVkAudio(server, message)
	if err != nil {
		return
	}

	err = sendMessage(fileId, ownerId, userId)
	if err != nil {
		return
	}
}

func sendMessage(fileId, ownerId, userId string) error {
	randomId := string(rand.Int() + rand.Intn(100000))
	document := fmt.Sprintf("doc%s_%s", fileId, ownerId)

	urlArgs := url.Values{}
	urlArgs.Add("user_id", userId)
	urlArgs.Add("random_id", randomId)
	urlArgs.Add("attachment", document)
	urlArgs.Add("access_token", accessToken)
	urlArgs.Add("v", v)
	urlInfo := urlArgs.Encode()

	fullUrl := ApiMessage + urlInfo

	res, err := http.Get(fullUrl)
	if err != nil {
		logger.Println("error in send message " + err.Error())
		return err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Println("error in json decode " + err.Error())
		return err
	}

	var resMessage messageApi

	err = json.Unmarshal(resBody, resMessage)
	if err != nil {
		logger.Println("error in json decode " + err.Error())
		return err
	}

	return nil
}

func uploadVkAudio(server, track string) (string, string, error) {
	file, err := os.Open(track)
	if err != nil {
		logger.Println("error when open file " + err.Error())
		return "", "", err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		logger.Println("error create form file " + err.Error())
		return "", "", err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		logger.Println("error file close " + err.Error())
		return "", "", err
	}

	err = writer.Close()
	if err != nil {
		logger.Println("error file close " + err.Error())
		return "", "", err
	}

	r, err := http.NewRequest("POST", server, body)
	if err != nil {
		logger.Println("error create request " + err.Error())
		return "", "", err
	}

	r.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}

	res, err := client.Do(r)
	if err != nil {
		logger.Println("error send request " + err.Error())
		return "", "", err
	}

	var uploadFile saveApi
	resBody, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal([]byte(resBody), &uploadFile)

	fileId, ownerId, err := saveFileVk(uploadFile.File)
	if err != nil {
		logger.Println("error on save file in vk " + err.Error())
		return "", "", err
	}

	err = os.Remove(track)
	if err != nil {
		logger.Println("error on remove file " + err.Error())
		return "", "", err
	}

	err = file.Close()
	if err != nil {
		logger.Println("error file close " + err.Error())
		return "", "", err
	}

	return fileId, ownerId, err
}

func saveFileVk(file string) (string, string, error) {
	urlArgs := url.Values{}
	urlArgs.Add("file", file)
	urlArgs.Add("access_token", accessToken)
	urlArgs.Add("v", v)
	urlInfo := urlArgs.Encode()

	fullUrl := ApiSave + urlInfo

	res, err := http.Get(fullUrl)
	if err != nil {
		logger.Println("error on send request " + err.Error())
		return "", "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	var saveRes saveResponseApi

	err = json.Unmarshal(body, &saveRes)
	if err != nil {
		logger.Println("error on decode json saveResponse " + err.Error())
		return "", "", err
	}

	return saveRes.Id, saveRes.OwnerId, nil
}

func getVkUploadServer(peerId string) (string, error) {
	var server uploadApi
	baseUrl := ApiUploadServer

	urlArgs := url.Values{}
	urlArgs.Add("type", "audio_message")
	urlArgs.Add("peer_id", peerId)
	urlArgs.Add("access_token", accessToken)
	urlArgs.Add("v", v)
	urlInfo := urlArgs.Encode()

	Url := baseUrl + urlInfo

	resp, err := http.Get(Url)
	if err != nil {
		logger.Println("error in request vk " + err.Error())
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Println("error in read response vk " + err.Error())
		return "", err
	}

	err = json.Unmarshal(body, &server)
	if err != nil {
		logger.Println("error on unmarshal json " + err.Error())
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
		logger.Println(`can't connect to redis ` + err.Error())
		return false
	}

	logger.Println("connected to redis")

	return true
}

func logging() {
	var err error
	logFile, err = os.OpenFile("var/log/sendMail.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error create log file " + err.Error())
		os.Exit(1)
	}

	logger = log.New(logFile, "", log.LstdFlags)
}

func destruct() {
	_ = queue.Close()
	_ = logFile.Close()
}
