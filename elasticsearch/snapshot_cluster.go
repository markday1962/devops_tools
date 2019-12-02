// Build for AWS EC2
//
// GOOS=linux GOARCH=amd64 go build -o es-snapshot main.go
//
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"gopkg.in/olivere/elastic.v6"
)

type cluster struct {
	index    string
	name     string
	repo     string
	snapshot string
}

func main() {

	r := ""
	if len(os.Args) == 1 {
		r = "cipher"
	} else if len(os.Args) == 2 {
		r = os.Args[1]
	} else {
		fmt.Println("Usage: pass zero or one elasticsearch repo")
	}
	fmt.Printf("Elasticsearch repo %v is being used for snapshot\n", r)

	t := time.Now()
	dt := t.Format("2006-01-02-150405")

	cl := cluster{
		repo:     r,
		snapshot: "cipher-snapshot-" + dt,
	}

	// Use NewClient to make a connection to the cluster
	url := "http://localhost:9200"
	client, err := elastic.NewClient(elastic.SetURL(url))
	ctx := context.Background()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Use ClusterHealth to check the cluster health
	fmt.Printf("Retrieving cluster health from %v\n", url)
	response, err := client.ClusterHealth().Do(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	if response == nil {
		panic(err)
	}

	if response.Status != "green" {
		fmt.Printf("Cluster health of %v is currently %v\n",
			response.ClusterName, response.Status)
		fmt.Println("Error: cluster health must be green before a" +
			" snapshot can take place, Exiting")
		return
	}
	if response.Status == "green" {
		fmt.Printf("Cluster health of %v is currently %v\n",
			response.ClusterName, response.Status)
	}

	//Use SnapshotCreate to snapshot all indices on a cluster
	snapshot, err := client.SnapshotCreate(cl.repo, cl.snapshot).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	if snapshot == nil {
		fmt.Println("Error: Response to snapshot request is empty, " +
			"please investigate, Exiting")
		return
	}
	if snapshot != nil {
		fmt.Printf("Snapshot %v for cluster %v is being created\n",
			cl.snapshot, response.ClusterName)
	}
}
