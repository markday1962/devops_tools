package main


import (
	"fmt"
	"os"
	"github.com/aws/aws-sdk-go/aws"
	_"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Bucket, item\nUsage: go run bucket item")
		os.Exit(2)
	}
}

func makeSession() *session.Session {
	// Specify profile to load for the session's config
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2"),
	})
	if err != nil {
		fmt.Println("failed to create session,", err)
		fmt.Println(err)
		os.Exit(2)
	}
}

func getObject(host string, fp string) {



	//sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-west-2")})
	//sess, err := session.NewSession()

	sess := session.New()

	svc := s3.New(sess)
}
