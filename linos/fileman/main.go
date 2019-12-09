package main

import (
	"fmt"
	"log/syslog"
	"os"
)

func main() {

	//comparable files
	var comparables = []string{"assignee_to_eset.bin", "eset_to_size.bin", "patfam_to_assignee.bin", "patfam_to_patfam.bin"}

	// Create a new syslog object
	logger, err :=syslog.New(syslog.LOG_LOCAL3, "aist-fileman")
	if err != nil {
		fmt.Println(err)
	}

	defer logger.Close()
	_ = logger.Notice("Starting Redis BGSAVE")

	// Comparables directory
	directory := "/mnt/data/comparables-service-cache/"
	//directory := "/tmp/"

	// Open the directory and read all its files.
	dirRead, _ := os.Open(directory)
	dirFiles, _ := dirRead.Readdir(0)

	// Loop over the directory's files.
	for index := range(dirFiles) {
		filePath := dirFiles[index]

		// Get name of file and its full path.
		fileName := filePath.Name()

		// Check to see if file is in comparable slice
		v := Find(comparables, fileName)

		if v {
			fullPath := directory + fileName
			os.Remove(fullPath)
			logger.Notice("Removed file: " + fullPath)
		}
	}
}

func Find(a []string, b string) bool {
	// Determine if a specified item is present in a slice
	for _, i := range a{
		if i == b {
			return true
		}
	}
	return false
}