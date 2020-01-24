// parked up currently getting the following error Error received: unknown path '/_admin/backup/create'
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
	"path/filepath"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

const bpath = "/tmp/arangodb"

func main() {
	// Create a new syslog object
	logger, err :=syslog.New(syslog.LOG_LOCAL3, "aist-arangodb")
	if err != nil {
		fmt.Println(err)
	}
	defer logger.Close()
	backupdir := createLogDir()
	backupArrangodb(backupdir, logger)
}

func createLogDir() string {
	// Create a new backup folder for the arrangodb
	currentTime := time.Now()
	newfolder := currentTime.Format("20060107")
	newpath := filepath.Join(bpath, newfolder)

	if _, err := os.Stat(newpath); os.IsNotExist(err){
		err := os.MkdirAll(newpath, os.ModePerm)
		if err != nil {

		}
	}
	return newpath
}

func removeOldBackups(){
	// Remove previous backup folders
	files, err := ioutil.ReadDir(bpath)
	if err != nil {
		log.Fatal(err)
	}

	var t = time.Now().String()
	fmt.Printf(t)

	for _, f := range files {
		if f.IsDir(){
			fmt.Printf("%T\n",f.ModTime())
		}
	}
}

type BackupCreateOptions struct {
	Label             string
	AllowInconsistent bool
	Timeout           time.Duration
}

func newBackupCreateOptions() *BackupCreateOptions{
	bco := BackupCreateOptions{Label: "Test", AllowInconsistent: false, Timeout: 300}
	return &bco
}

func backupArrangodb(backupdir string, logger *syslog.Writer){

	_ = logger.Notice("Starting to backup arrangodb")

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:30529"},
	})
	if err != nil {
		// Handle error
	}

	c, err := driver.NewClient(driver.ClientConfig{
		Connection: conn,
	})
	if err != nil {
		// Handle error
	}

	ctx := context.Background()
	opt := newBackupCreateOptions()

	id, resp, err := c.Backup().Create(ctx, (*driver.BackupCreateOptions)(opt))
	if err != nil {
		fmt.Printf("Error received: %v", err.Error())
	}

	fmt.Printf("\nBackup id: %v", id)
	fmt.Printf("\nBackup response: %v", resp)
}
