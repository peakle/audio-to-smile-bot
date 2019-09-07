package MailDemon

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

type QueueBody struct {
	UserId  string `json:"user_id"`
	Message string `json:"message"`
}

type VkServer struct {
	UploadUrl string `json:"upload_url"`
}

var queue *redis.Client
var consumerCount uint8 = 0
var mu sync.Mutex

const sendQ = "queue_send"

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("error when open env")
		return
	}

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

	task := queue.Get(sendQ)
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
	sendVkAudio(server, body)
}

func sendVkAudio(server, trackname string) {
	//TODO
}

func getVkUploadServer(peerId string) string {
	var server VkServer
	baseUrl := os.Getenv("VK_SERVER")
	ver := os.Getenv("VK_API_VERSION")
	url := baseUrl + peerId + "&v=" + ver

	resp, err := http.Get(url)

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
