package main

import (
	"fmt"
	"time"

	"gopkg.in/olivere/elastic.v3"
)

func clusterHealth() {
	// make a connection
	url := "http://livees.prod.aistemos.com:9200"
	client, err := elastic.NewClient(elastic.SetURL(url))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	response, err := client.ClusterHealth().Do()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	if response.Status == "green" {
		fmt.Printf("Cluster health is %s\n", response.Status)
	} else {
		fmt.Printf("Cluster health is %s investigate\n", response.Status)
	}
}

func clusterPing() {
	// make a connection
	url := "http://livees.prod.aistemos.com:9200"
	client, err := elastic.NewClient(elastic.SetURL(url))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	_, code, err := client.Ping(url).Do()
	fmt.Printf("Response code received %d\n", code)

}

func main() {
	fmt.Println("Getting cluster health")
	for {
		clusterHealth()
		clusterPing()
		time.Sleep(10 * time.Second)
	}
}
