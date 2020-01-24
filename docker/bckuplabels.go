package main

import (
	"context"
	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main(){
	if len(os.Args) != 3 {
		usage()
	}
	bucket := os.Args[1]
	pk := os.Args[2]
	id := getid()
	n := nodeinspect()
	fp := writedata(id, n)
	dirtycopy(bucket,pk, fp)
}

type Node struct {
	ID string
	Name string
	Labels []Label
}

type Label struct {
	Key string
	Value string
}

func usage(){
	fmt.Println("Incorrect usage....\n")
	fmt.Println("Usage:\n")
	fmt.Println("bckuplbl <bucket> <prekey>")
	fmt.Println("bckuplbl mybucket asubfolder/anotherfolder")
	os.Exit(1)
}

func getid() string {
	//get's the swarm id
	cli, err := docker.NewEnvClient()
	if err != nil {
		panic(err)
	}
	swarm, err := cli.SwarmInspect(context.Background(),)
	if err != nil {
		panic(err)
	}
	id := swarm.ID
	return id
}

func nodeinspect() []Node {
	// Inspects the nodes in a cluster for their labels
	nodedata := []Node{}
	cli, err := docker.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// retrieve the nodes running in the swarm
	nodes, err := cli.NodeList(context.Background(), types.NodeListOptions{})
	if err != nil {
		panic(err)
	}

	// get the labels from the node
	for _, node := range nodes {
		data := Node {
			ID: node.ID,
			Name: node.Description.Hostname,
		}
		//get the nodes labels
		labels := node.Spec.Annotations
		for k, v := range labels.Labels{
			ldata := Label{
				Key: k,
				Value: v,
			}
			data.Labels = append(data.Labels,ldata)
		}
		nodedata = append(nodedata,data )
	}
	return nodedata
}

func writedata(id string, nodes []Node) string {
	// write nodes labeling to a file
	fpath := "/tmp/" + id + "-nodes-labels.txt"
	f, err := os.Create(fpath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write([]byte("Cluster Details:\n"))
	cluster := "ID:" + id + "\n"
	f.Write([]byte(cluster))
	for _, v := range nodes {
		f.Write([]byte("Node Details:\n"))
		//fmt.Printf("%v:%v\n",v.ID,v.Name)
		header := "Node:" + v.ID + ":" + v.Name + "\n"
		f.Write([]byte(header))
		f.Write([]byte("Node Labels:\n"))
		for _, v := range v.Labels{
			line := v.Key + ":" + v.Value + "\n"
			f.Write([]byte(line))
		}

		f.Write([]byte("\n"))
	}

	return fpath
}

func dirtycopy(bucket string, pk string, fp string){
	//quick dirty copy to s3 bucket using the aws cli installed on the host
	region := "eu-west-1"
	bucket = "s3://" + bucket + "/" + pk + "/"
	cmd := exec.Command("aws","s3", "cp", "--region", region, fp, bucket)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}