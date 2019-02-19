package service

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mingrammer/cfmt"
	"github.com/mingrammer/dynamodb-toolkit/config"
)

// NewDynamoDBClient creates a dynamodb client
func NewDynamoDBClient() (*dynamodb.DynamoDB, error) {
	awsConf := config.GetAWSConfig()
	profile := config.GetProfile()
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            *awsConf,
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, errors.New(cfmt.Serror(err.Error()))
	}
	return dynamodb.New(sess), nil
}
