package toolkit

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mingrammer/dynamodb-toolkit/mock"
)

func TestTruncate(t *testing.T) {
	dummySize := 1000
	deleteChunk := 25

	client := mock.NewDynamoDBClient()
	truncator := NewTruncator(client)

	// Create a table
	name := "user"
	truncator.client.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: aws.String("N"),
			},
		},
		BillingMode: aws.String("PRIVISIONED"),
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(name),
	})

	// Insert some test data
	req := []*dynamodb.WriteRequest{}
	for i := 0; i < dummySize; i++ {
		req = append(req, &dynamodb.WriteRequest{
			PutRequest: &dynamodb.PutRequest{
				Item: map[string]*dynamodb.AttributeValue{
					"id": {
						N: aws.String(strconv.Itoa(i + 1)),
					},
				},
			},
		})
		if (i+1)%deleteChunk == 0 || i >= dummySize-1 {
			client.BatchWriteItem(&dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]*dynamodb.WriteRequest{
					name: req,
				},
			})
			req = []*dynamodb.WriteRequest{}
		}
	}

	// Truncate
	if errs := truncator.Truncate([]string{name}, false); len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("There should be no errors, Got %s\n", err.Error())
		}
	}

	// Check the number of items is zero
	desc, err := client.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(name),
	})
	if err != nil {
		t.Errorf("There should be no errors, Got %s\n", err.Error())
	}
	if *desc.Table.ItemCount > 0 {
		t.Errorf("There should be no items, %d items is remaining\n", *desc.Table.ItemCount)
	}
}

func TestTruncateWithRecreate(t *testing.T) {
	dummySize := 1000
	deleteChunk := 25

	client := mock.NewDynamoDBClient()
	truncator := NewTruncator(client)

	// Create a table
	name := "item"
	output, _ := truncator.client.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("user_id"),
				AttributeType: aws.String("N"),
			},
			{
				AttributeName: aws.String("item_id"),
				AttributeType: aws.String("N"),
			},
		},
		BillingMode: aws.String("PAY_PER_REQUEST"),
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("user_id"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("item_id"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(name),
	})

	// Insert some test data
	req := []*dynamodb.WriteRequest{}
	for i := 0; i < dummySize; i++ {
		req = append(req, &dynamodb.WriteRequest{
			PutRequest: &dynamodb.PutRequest{
				Item: map[string]*dynamodb.AttributeValue{
					"user_id": {
						N: aws.String(strconv.Itoa(i + 1)),
					},
					"item_id": {
						N: aws.String(strconv.Itoa(rand.Intn(10))),
					},
				},
			},
		})
		if (i+1)%deleteChunk == 0 || i >= dummySize-1 {
			client.BatchWriteItem(&dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]*dynamodb.WriteRequest{
					name: req,
				},
			})
			req = []*dynamodb.WriteRequest{}
		}
	}

	// Truncate with recreate option
	if errs := truncator.Truncate([]string{name}, true); len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("There should be no errors, Got %s\n", err.Error())
		}
	}

	// Check the number of items is zero
	desc, err := client.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(name),
	})
	if err != nil {
		t.Errorf("There should be no errors, Got %s\n", err.Error())
	}
	if *desc.Table.ItemCount > 0 {
		t.Errorf("There should be no items, %d items is remaining\n", *desc.Table.ItemCount)
	}
	if (*desc.Table.CreationDateTime).Before(*output.TableDescription.CreationDateTime) {
		t.Errorf("Creation datetime of the recreated table should be after old one\n")
	}
}
