package main

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"net/http"
	"os"
)

const (
	QueueCreate  = "queue_create"
	Confirmation = "confirmation"
	MessageNew   = "message_new"
)

var (
	secret              string
	groupId             string
	vkConfirmationToken string
	queue               *redis.Client
	err                 error
)

type vkMessage struct {
	Secret  string            `json:"secret"`
	GroupId string            `json:"group_id"`
	Type    string            `json:"type"`
	Object  map[string]string `json:"object"`
}

type vkOutMessage struct {
	UserId  string `json:"user_id"`
	Message string `json:"message"`
}

func main() {
	err = godotenv.Load(".env")
	if err != nil {
		os.Exit(1)
	}
	if !redisConnect() {
		os.Exit(1)
	}
	defer destruct()

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		var body []byte
		_, _ = request.Body.Read(body)

		message := vkMessage{}
		_ = json.Unmarshal(body, &message)

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

			queueBody, _ := json.Marshal(vkOutMessage{
				UserId:  message.Object["from_id"],
				Message: message.Object["message"],
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
		Password: "",
		DB:       0,
	})

	res, err := queue.Ping().Result()

	if err != nil || res == "" {
		return false
	}

	return true
}

func destruct() {
	_ = queue.Close()
}
