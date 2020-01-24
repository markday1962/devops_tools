package main
// A simple app to save redis databases before flushing them
// this is very destructive
import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"log/syslog"
)

func main() {
	lc := liveCipher()
	v := checkhost(lc)
	fmt.Printf("%v", v)
	manageRedisNode()
}

func manageRedisNode() {
	// Create a new syslog object
	logger, err :=syslog.New(syslog.LOG_LOCAL3, "aist-bgsave")
	if err != nil {
		fmt.Println(err)
	}

	// Create redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
	})

	// Test redis connection
	pong, err := rdb.Ping().Result()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(pong)

	// Get last save in unix time
	ls := rdb.LastSave()
	lst, err := ls.Result()
	if err != nil {
		fmt.Println(err)
		_ = logger.Notice("Error retrieving LastSave value")
		os.Exit(1)
	}
	msg := "Current LastSave value: " + strconv.FormatInt(lst, 10)
	_ = logger.Notice(msg)
	fmt.Printf("%v\n", msg)

	// start background save
	resp := rdb.BgSave()
	fmt.Println(resp)
	time.Sleep(2 * time.Second)
	nls := rdb.LastSave()
	nlst, err := nls.Result()
	if err != nil {
		fmt.Println(err)
		_ = logger.Notice("Error retrieving results of BGSAVE")
		os.Exit(1)
	}

	for {
		if nlst > lst {
			msg = "New LastSave value: " + strconv.FormatInt(nlst, 10)
			_ = logger.Notice(msg)
			logger.Notice("Redis BGSAVE has completed.")
			fmt.Println(msg)
			fmt.Println("Redis BGSAVE has completed.")
			break
		}
		time.Sleep(30 * time.Second)
		nls = rdb.LastSave()
		nlst, err = nls.Result()
		if err != nil {
			fmt.Println(err)
			_ = logger.Notice("Error retrieving new results of BGSAVE")
			os.Exit(1)
		}
	}

	// Flush all redis database on host
	//logger.Notice("Flushing redis database.")
	//resp = rdb.FlushAll()
	//fmt.Println(resp)
}

func liveCipher() string {
	// Gets the current livecipher
	url := "http://app.cipher.ai/version"

	req, _ := http.NewRequest("GET", url, nil)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var decode []interface{}
	err := json.Unmarshal([]byte(body), &decode)
	if err != nil {
		err.Error()
	}

	app := decode[0]
	m := app.(map[string]interface{})
	for k, v := range m {
		if k == "Aistemos-Application-Id" {
			out, _ := json.Marshal(v)
			// check and remove extra double quotes if they are there
			if len(out) > 0 && out[0] == '"' {
				out = out[1:]
			}
			if len(out) > 0 && out[len(out)-1] == '"' {
				out = out[:len(out)-1]
			}
			return string(out)
		}
	}
	return ""
}

func checkhost(lc string) bool {
	// Checks the host running the code does not have a livecipher hostname
	h, err := os.Hostname()
	if err != nil {
		err.Error()
	}
	return strings.Contains(h, lc)
}

