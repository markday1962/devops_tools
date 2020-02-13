package main

// A simple app to save redis databases
import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	_ "github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-redis/redis"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) != 4 {
		usage()
	}
	saveRedis()
	bucket := os.Args[1]
	prekey := os.Args[2]
	f := os.Args[3]
	fp := findFile("/mnt", f)

	// copyFile(bucket, prekey, fp, f)
	dirtyAWSCopy(bucket, prekey, fp, f)

}

func usage(){
	fmt.Println("Incorrect usage:")
	fmt.Println("sac <bucket> <prefix> <file-to-upload>\nsac aistemos-data-backups redis/devcipher-data dump.rdb")
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

func dirtyAWSCopy(bucket string, prekey string, fp string, f string){
	os.Chdir(fp)
	region := "eu-west-1"
	bucket = "s3://" + bucket + "/" + prekey + "/"
	fmt.Printf("%v\n", bucket)
	cmd := exec.Command("aws","s3", "cp", "--region", region, fp, bucket)
	err:= cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func copyFile(bucket string, prekey string, fp string, f string) {
	// Copy file to S3
	fmt.Printf("Printing file path %v \n", fp)
	os.Chdir(fp)
	file, err := os.Open(fp)
	if err != nil {
		fmt.Printf("Unable to open file: %v\n", err)
		os.Exit(1)
	}
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	defer file.Close()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2")},
	)
	uploader := s3manager.NewUploader(sess)

	t := time.Now()
	dt := t.Format("20060102")
	key := "/" + prekey + "/" + f + "-" + dt

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key: aws.String(key),
		Body: file,
	})
	if err != nil {
		// Print the error and exit.
		fmt.Printf("Unable to upload %q to %q, %v\n", fp, bucket, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully uploaded %q to %q\n", fp, bucket + key)

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
