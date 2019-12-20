package main

// A simple app to save redis databases
import (
	"fmt"
	"log/syslog"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func main() {
	backupRedis()
}

func backupRedis() {
	// Create a new syslog object
	logger, err :=syslog.New(syslog.LOG_LOCAL3, "aist-bgsave")
	if err != nil {
		fmt.Println(err)
	}

	defer logger.Close()
	_ = logger.Notice("Starting Redis BGSAVE")

	// Create a new redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // use default Addr
	})

	defer rdb.Close()

	// Test redis connection with PING
	pong, err := rdb.Ping().Result()
	if err != nil {
		fmt.Println(err)
		_ = logger.Notice("Error to Redis PING")
		os.Exit(1)
	}
	msg := "Response to PING: " + pong
	fmt.Printf("%v\n", msg)
	_ = logger.Notice(msg)

	// Get last save in unix time
	ls := rdb.LastSave()
	lst, err := ls.Result()
	if err != nil {
		fmt.Println(err)
		_ = logger.Notice("Error retrieving LastSave value")
		os.Exit(1)
	}
	msg = "Current LastSave value: " + strconv.FormatInt(lst, 10)
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
}
