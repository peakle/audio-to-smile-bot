package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"html"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"
)

type Message struct {
	Track       string `json:"track"`
	UserId      string `json:"user_id"`
	MessageBody string `json:"message"`
}

var (
	queue         *redis.Client
	err           error
	logFile       *os.File
	logger        *log.Logger
	workdir       string
	consumerCount uint8 = 0
	mu            sync.Mutex
)

const CreateQ = "queue_create"
const SendQ = "queue_send"

func main() {
	logging()
	err = godotenv.Load(".env")

	workdir = os.Getenv("WORKDIR")

	if err != nil {
		logger.Println("error when open env")
		return
	}

	isRedis := redisConnect()
	defer destruct()

	if isRedis {
		handle()
	}
}

func handle() {
	for {
		queueLen := queue.LLen(CreateQ).Val()
		if queueLen > 0 && consumerCount < 2 {
			task := queue.LPop(CreateQ)

			taskBody, _ := task.Result()

			if len(taskBody) > 4 {
				go consumer(taskBody)
				consumerCount++
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func consumer(task string) {
	var queueBody Message

	defer closeConsumer(task)

	err = json.Unmarshal([]byte(task), &queueBody)
	if err != nil {
		logger.Println("error in decode json", err.Error())
		return
	}

	body := queueBody.MessageBody
	userId := queueBody.UserId

	emojiList := findEmoji(body)

	if len(emojiList) == 0 {
		return
	}

	var track string
	track, err = generateTrack(emojiList)
	if err != nil || len(track) == 0 {
		return
	}

	var messageBody string
	for _, emoji := range emojiList {
		messageBody += html.UnescapeString("&#"+strconv.Itoa(emoji)+";") + " "
	}

	var message []byte
	message, err = json.Marshal(Message{
		Track:       track,
		UserId:      userId,
		MessageBody: messageBody,
	})

	if err != nil {
		logger.Println("error on create message", err.Error())
		return
	}

	mu.Lock()
	queue.RPush(SendQ, message, 0)
	mu.Unlock()
}

func closeConsumer(task string) {
	if err != nil {
		queue.RPush(CreateQ, task)
		//TODO error counter for corrupted tasks
		log.Println("task with error", task)
	}
	consumerCount--
}

func generateTrack(emojiList []int) (string, error) {
	var err error
	randName := rand.Int() + rand.Intn(1000)
	newTrack := workdir + "/mails/" + strconv.Itoa(randName) + ".ogg"
	var sampleList []string
	var sample string
	var isSet bool

	for num, code := range emojiList {
		sample, isSet = emojiMap[code]
		if !isSet {
			continue
		}

		sampleList = append(sampleList, workdir+"/samples/"+sample)

		if num > 10 {
			break
		}
	}

	if len(sampleList) > 0 {
		command := "sox"

		if len(sampleList) == 1 {
			command = "cp"
		}

		sampleList = append(sampleList, newTrack)
		cmd := exec.Command(command, sampleList...)
		err = cmd.Run()
		if err != nil {
			logger.Println("error on generate track", err.Error())
			return "", err
		}

		return newTrack, nil
	}

	return "", err
}

func findEmoji(text string) []int {
	var results []int
	re := regexp.MustCompile(`[\x{2700}-\x{27BF}]|[\x{2600}-\x{26FF}]|[\x{1D100}-\x{1D1FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{1F900}-\x{1F9FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F300}-\x{1F5FF}]|([0-9#][\x{20E3}])|[\x{00ae}\x{00a9}\x{203C}\x{2047}\x{2048}\x{2049}\x{3030}\x{303D}\x{2139}\x{2122}\x{3297}\x{3299}][\x{FE00}-\x{FEFF}]?|[\x{2190}-\x{21FF}][\x{FE00}-\x{FEFF}]?|[\x{2300}-\x{23FF}][\x{FE00}-\x{FEFF}]?|[\x{2460}-\x{24FF}][\x{FE00}-\x{FEFF}]?|[\x{25A0}-\x{25FF}][\x{FE00}-\x{FEFF}]?|[\x{2600}-\x{27BF}][\x{FE00}-\x{FEFF}]?|[\x{2900}-\x{297F}][\x{FE00}-\x{FEFF}]?|[\x{2B00}-\x{2BF0}][\x{FE00}-\x{FEFF}]?|[\x{1F000}-\x{1F6FF}][\x{FE00}-\x{FEFF}]?`)
	if len(re.FindStringIndex(text)) > 0 {
		emoji := re.FindAllString(text, -1)

		for _, n := range emoji {
			code, _ := utf8.DecodeRuneInString(n)
			results = append(results, int(code))
		}
	}

	return results
}

func redisConnect() bool {
	redisAddr := os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
	redisPass := os.Getenv("REDIS_PASSWORD")

	queue = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       0,
	})

	res, err := queue.Ping().Result()

	if err != nil || res == "" {
		logger.Println("can't connect to redis")
		return false
	}

	logger.Println("connected to redis")

	return true
}

func logging() {
	logFile, err = os.OpenFile("var/log/createMessage.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error on open create log " + err.Error())
		os.Exit(1)
	}

	logger = log.New(logFile, "", log.LstdFlags)
}

func destruct() {
	_ = queue.Close()
	_ = logFile.Close()
}
