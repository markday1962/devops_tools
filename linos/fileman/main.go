package main

import (
	"fmt"
	"log/syslog"
	"os"
)

func main() {

	// Create a new syslog object
	logger, err :=syslog.New(syslog.LOG_LOCAL3, "aist-fileman")
	if err != nil {
		fmt.Println(err)
	}

	defer logger.Close()
	_ = logger.Notice("Starting Redis BGSAVE")

	// The target directory.
	directory := "/home/sam/test/"

	// Open the directory and read all its files.
	dirRead, _ := os.Open(directory)
	dirFiles, _ := dirRead.Readdir(0)

	// Loop over the directory's files.
	for index := range(dirFiles) {
		fileHere := dirFiles[index]

		// Get name of file and its full path.
		nameHere := fileHere.Name()
		fullPath := directory + nameHere

		// Remove the file.
		os.Remove(fullPath)
		fmt.Println("Removed file:", fullPath)
	}
}
