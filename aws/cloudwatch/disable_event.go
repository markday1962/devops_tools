package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"os"
)

const DEFAULT_EVENT string = "aistStopDevelopmentInstances"

func main() {

	name := ""

	if len(os.Args) != 2 {
		fmt.Printf("Event name not supplied, using default off %v\n", DEFAULT_EVENT)
		name = DEFAULT_EVENT
	} else {
		name = os.Args[1]
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := cloudwatchevents.New(sess)


	resp, err := svc.DisableRule(&cloudwatchevents.DisableRuleInput{
		Name: aws.String(name),
	})

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Println(resp)
	fmt.Println("Success")
}
