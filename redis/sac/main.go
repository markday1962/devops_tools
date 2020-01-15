package main

// A simple app to save redis databases
import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	_ "github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-redis/redis"
	"log/syslog"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		usage()
	}
	saveRedis()
	host := os.Args[1]
	fp := findFile("/mnt", "dump.rdb")
	copyFile(host, fp)

}

func usage(){
	fmt.Println("Incorrect usage:")
	fmt.Println("sac <s3 key>\nsac redis-s3-folder")
	fmt.Println("Exiting...")
	os.Exit(1)
}

func saveRedis() {
	// Create a new syslog object
	logger, err := syslog.New(syslog.LOG_LOCAL3, "aist-bgsave")
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

func copyFile(host string, fp string) {
	//copies the dump.rdb to s3
	svc := s3.New(makeSession())
	bucket := "aistemos-data-backups"
	copySource := fp
	dt := time.Now().Format("20060102")
	key := "/redis/" + host + "/" + "dump.rdb-" + dt
	fmt.Printf("Copying %v to %v\n", fp, key)

	input := &s3.CopyObjectInput{
		Bucket: aws.String(bucket),
		CopySource: aws.String(copySource),
		Key: aws.String("test"),
	}

	result, err := svc.CopyObject(input)
	if err != nil {
		fmt.Printf(err.Error())
	}
	fmt.Println(result)
}

func makeSession() *session.Session {
	// Specify profile to load for the session's config

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2"),
	})
	if err != nil {
		fmt.Println("failed to create session,", err)
		fmt.Println(err)
		os.Exit(2)
	}
	return sess
}

func findFile(rootDir string, dumpFile string) string {
	// find the location of the dump.rdb file
	fpath := ""
	_ = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.Name() == dumpFile {
			fpath = path
		}

		return nil
	})
	return fpath
}

