package main
// A simple app to save redis databases before flushing them
// this is very destructive
import (
	"fmt"
	"github.com/go-redis/redis"
	"os"
	"time"
)

func main() {
	manageRedisNode()
}

func manageRedisNode() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // use default Addr
	})

	// Test connection
	pong, err := rdb.Ping().Result()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(pong)

	// Get last save in unix time
	ls := rdb.LastSave()
	fmt.Println(ls)
	resp :=rdb.BgSave()
	fmt.Println(resp)
	time.Sleep(2 * time.Second)
	nls := rdb.LastSave()

	for {
		if nls != ls {
			fmt.Println(nls)
			fmt.Println("Save completed")
			break
		}
		fmt.Println("Waiting for save to complete")
		time.Sleep(2 * time.Second)
		nls = rdb.LastSave()
	}

	// Flush all redis database on host
	//resp = rdb.FlushAll()
	//fmt.Println(resp)
}