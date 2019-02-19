package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// Config holds aws configuration and profile
type Config struct {
	awsConf *aws.Config
	profile string
}

var config *Config

func init() {
	config = &Config{awsConf: new(aws.Config)}
}

// GetAWSConfig returns aws configuration
func GetAWSConfig() *aws.Config {
	return config.awsConf
}

// GetProfile returns profile
func GetProfile() string {
	return config.profile
}

// SetCredentials sets the static aws credentials
func SetCredentials(accessKeyID, secretAccessKey string) {
	if accessKeyID != "" && secretAccessKey != "" {
		creds := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
		config.awsConf.WithCredentials(creds)
	}
}

// SetProfile sets the aws profile
func SetProfile(profile string) {
	if profile != "" {
		config.profile = profile
	}
}

// SetRegion sets the aws dynamodb region
func SetRegion(region string) {
	if region != "" {
		config.awsConf.WithRegion(region)
	}
}

// SetEndpoint sets the aws dynamodb endpoint
func SetEndpoint(endpoint string) {
	if endpoint != "" {
		config.awsConf.WithEndpoint(endpoint)
	}
}

// Reset resets the global configuration
func Reset() {
	config.awsConf = new(aws.Config)
	config.profile = ""
}
