package main

import (
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	"strings"
)

func getIndexes() {
	// make a connection
	var as = "assignee-search"
	url := "http://localhost:9200"
	client, err := elastic.NewClient(elastic.SetURL(url))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// get all ES indices
	response, err := client.IndexNames()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	for _, v := range response {
		// currently only deleting the assignee-search indices
		if strings.Contains(v, as){
			fmt.Printf("Preparing to delete Elasticsearch indices %v\n", v)
			deleteIndex, err := client.DeleteIndex(v).Do()
			if err != nil {
				fmt.Println(err)
				panic(err)
			} else {
				fmt.Printf("Acknowledgement to deleting %v: %v\n",v, deleteIndex.Acknowledged)
			}
		}
	}
}

func main() {
	getIndexes()
}
