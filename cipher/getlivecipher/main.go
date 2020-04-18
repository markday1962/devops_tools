package main

import (
	"encoding/json"
	"fmt"
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

func Handler() (string, error) {
	resp, err := http.Get("http://app.cipher.ai/version")
	if err != nil {
		fmt.Println("Error response too http request %v", err)
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf(err.Error())
		return "", err
	}
	x := body[1: len(body)-1]
	c := new(CipherResponse)
	err = json.Unmarshal([]byte(x), c)
	if err != nil{
		fmt.Printf(err.Error())
		return "", err
	}
	return c.Host, nil
}

func main() {
	lambda.Start(Handler)
}
