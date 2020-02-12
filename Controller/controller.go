package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"strconv"
)

const (
	QueueCreate  = "queue_create"
	Confirmation = "confirmation"
	MessageNew   = "message_new"
)

var (
	queue *redis.Client
	err   error
)

type vkMessage struct {
	Secret  string          `json:"secret"`
	GroupId int             `json:"group_id"`
	Type    string          `json:"type"`
	Object  vkMessageObject `json:"object"`
}

type vkMessageObject struct {
	FromId int    `json:"from_id"`
	Text   string `json:"text"`
}

type vkOutMessage struct {
	UserId  string `json:"user_id"`
	Message string `json:"message"`
}

func main() {
	err = godotenv.Load(".env")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if !redisConnect() {
		os.Exit(1)
	}
	defer destruct()

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		secret := os.Getenv("VK_API_SECRET")
		groupId, _ := strconv.Atoi(os.Getenv("VK_GROUP_ID"))
		vkConfirmationToken := os.Getenv("VK_CONFIRMATION_TOKEN")

		message := vkMessage{}
		err = json.NewDecoder(request.Body).Decode(&message)
		if err != nil {
			return
		}

		if message.Secret != secret || len(message.Secret) == 0 {
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte{})
			return
		}

		switch message.Type {
		case Confirmation:
			if message.GroupId != groupId {
				writer.WriteHeader(200)
				_, _ = writer.Write([]byte{})
				return
			}

			writer.WriteHeader(200)
			_, _ = writer.Write([]byte(vkConfirmationToken))

		case MessageNew:
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte("ok"))

			UserId := strconv.Itoa(message.Object.FromId)

			queueBody, _ := json.Marshal(vkOutMessage{
				UserId:  UserId,
				Message: message.Object.Text,
			})

			queue.RPush(QueueCreate, queueBody)

		default:
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte{})
		}
	})

	_ = http.ListenAndServe(":80", nil)
}

func redisConnect() bool {
	redisAddr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")

	queue = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	res, err := queue.Ping().Result()

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if res == "" {
		fmt.Println("redis not available")
		return false
	}

	return true
}

func destruct() {
	_ = queue.Close()
}
