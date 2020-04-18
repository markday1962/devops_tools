package main

import (
	"encoding/json"
	"fmt"
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"io/ioutil"
	"net/http"
)

type CipherResponse struct {
	FEVersion string `json:"Aistemos-FrontendService-Version"`
	Version string `json:"Aistemos-Software-Version"`
	DataVersion int64 `json:"Aistemos-Data-Version"`
	Host string `json:"Aistemos-Application-Id"`
}

type LambdaInputEvent struct {
	Host string `json:"host"`
}

func Handler(ctx context.Context, event LambdaInputEvent) (string, error) {
	fmt.Printf("Green Cipher host: %v\n", event.Host)
	resp, err := http.Get("http://app.cipher.ai/version")
	if err != nil {
		fmt.Printf("Error response too http request %v:\n", err)
		return "error", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Read http response error: %v:\n",err.Error())
		return "error", err
	}

	x := body[1: len(body)-1]
	c := new(CipherResponse)
	err = json.Unmarshal([]byte(x), c)
	if err != nil{
		fmt.Printf("Cipher response json unmarshall error %v:\n",err.Error())
		return "error", err
	}
	fmt.Printf("Blue Cipher host: %v\n", c.Host)

	if c.Host != event.Host {
		fmt.Printf("%v is not livecipher\n", event.Host)
		return "false", nil
	} else {
		fmt.Printf("%v is livecipher\n", event.Host)
		return "true", nil
	}
}

func main() {
	//host := "{\"host\": \"marvin\"}"
	//Handler(host)
	lambda.Start(Handler)
}
