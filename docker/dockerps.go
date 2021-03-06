package main

import (
	"context"
	"fmt"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
)

func main(){
	dockerps()
}

func dockerps() {
	cli, err := docker.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:20], container.Image)
	}
}