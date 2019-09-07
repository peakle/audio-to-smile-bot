package MessageDemon

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"
)

type QueueBody struct {
	UserId  string `json:"user_id"`
	Message string `json:"message"`
}

type Sample struct {
	Name string `json:"sample"`
}

var queue *redis.Client
var db *sql.DB

var (
	consumerCount uint8 = 0
	mu            sync.Mutex
)

const createQ = "queue_create"
const sendQ = "queue_send"

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("error when open env")
		return
	}

	isRedis := redisConnect()
	isMysql := mysqlConnect()

	if isMysql && isRedis {
		handle()
	} else {
		fmt.Println(`can't connect to redis`)
	}
}

func handle() {
	for {
		if consumerCount < 10 {
			consumerCount++
			go consumer()
		}

		time.Sleep(1)
	}
}

func consumer() {
	defer closeConsumer()

	var queueBody []QueueBody
	task := queue.Get(createQ)

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

	body := queueBody[0].Message
	userId := queueBody[0].UserId

	emojiList := findEmoji(body)

	track := generateTrack(emojiList)

	if len(track) == 0 {
		fmt.Println("error on create message")
		return
	}

	message, err := json.Marshal(QueueBody{
		Message: track,
		UserId:  userId,
	})

	if err != nil {
		fmt.Println("error on create message")
		return
	}

	queue.Set(sendQ, message, 0)
}

func closeConsumer() {
	mu.Lock()
	consumerCount--
	mu.Unlock()
}

func generateTrack(emojiList []int) string {
	randName := rand.Int() + rand.Intn(1000)
	full := strconv.Itoa(randName) + ".ogg"
	var tracks string

	for num, code := range emojiList {
		var sample Sample
		err := db.QueryRow("SELECT s.sample as sample FROM Smile as s WHERE s.code = ?", code).Scan(&sample.Name)

		if err != nil {
			fmt.Println("error in query mysql")
			return ""
		}

		tracks = tracks + sample.Name + ".ogg "

		if num > 10 {
			break
		}
	}

	if len(tracks) > 0 {
		cmd := exec.Command("sox", tracks+full)
		err := cmd.Run()

		if err != nil {
			fmt.Println("error on generate track")
			return ""
		}

		return full
	}

	return ""
}

func findEmoji(text string) []int {
	var results []int
	var re = regexp.MustCompile(`[\x{2700}-\x{27BF}]|[\x{2600}-\x{26FF}]|[\x{1D100}-\x{1D1FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{1F900}-\x{1F9FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F300}-\x{1F5FF}]|([0-9#][\x{20E3}])|[\x{00ae}\x{00a9}\x{203C}\x{2047}\x{2048}\x{2049}\x{3030}\x{303D}\x{2139}\x{2122}\x{3297}\x{3299}][\x{FE00}-\x{FEFF}]?|[\x{2190}-\x{21FF}][\x{FE00}-\x{FEFF}]?|[\x{2300}-\x{23FF}][\x{FE00}-\x{FEFF}]?|[\x{2460}-\x{24FF}][\x{FE00}-\x{FEFF}]?|[\x{25A0}-\x{25FF}][\x{FE00}-\x{FEFF}]?|[\x{2600}-\x{27BF}][\x{FE00}-\x{FEFF}]?|[\x{2900}-\x{297F}][\x{FE00}-\x{FEFF}]?|[\x{2B00}-\x{2BF0}][\x{FE00}-\x{FEFF}]?|[\x{1F000}-\x{1F6FF}][\x{FE00}-\x{FEFF}]?`)
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

func mysqlConnect() bool {
	var err error

	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	dbName := os.Getenv("MYSQL_DATABASE")

	cred := user + ":" + password + "@tcp(localhost:3306)/" + dbName

	db, err = sql.Open("mysql", cred)

	if err != nil {
		fmt.Println("error in connect to database")
		return false
	}

	return true
}
