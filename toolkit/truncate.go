package toolkit

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/mingrammer/cfmt"
	"github.com/mingrammer/dynamodb-toolkit/calc"
	"github.com/mingrammer/dynamodb-toolkit/retryer"
)

// Truncator holds dynamodb client
type Truncator struct {
	client dynamodbiface.DynamoDBAPI
}

const (
	megabyte = 1 << 20

	maxTotalSegments = 1000000
	deleteChunk      = 25
)

// NewTruncator creates a session and a dynamodb client
func NewTruncator(client dynamodbiface.DynamoDBAPI) *Truncator {
	return &Truncator{client: client}
}

func (t *Truncator) readMeta(table string) (*dynamodb.DescribeTableOutput, error) {
	meta, err := t.client.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(table),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeResourceNotFoundException:
				return nil, fmt.Errorf("Table '%s' is not found", table)
			}
		}
		return nil, fmt.Errorf("Something gone wrong while describing the table ,got %s", err.Error())
	}
	return meta, nil
}

func (t *Truncator) delete(table string, scanned *dynamodb.ScanOutput) error {
	errc := make(chan error, 1)
	wg := sync.WaitGroup{}
	wg.Add(int(math.Ceil(float64(len(scanned.Items)) / float64(deleteChunk))))
	req := []*dynamodb.WriteRequest{}
	for i, a := range scanned.Items {
		req = append(req, &dynamodb.WriteRequest{
			DeleteRequest: &dynamodb.DeleteRequest{
				Key: a,
			},
		})
		if (i+1)%deleteChunk == 0 || i >= int(*scanned.Count)-1 {
			go func(reqChunk []*dynamodb.WriteRequest) {
				defer wg.Done()
				unprocessed := map[string][]*dynamodb.WriteRequest{
					table: reqChunk,
				}
				attempts := 0
				for len(unprocessed[table]) > 0 {
					if attempts > 0 {
						time.Sleep(retryer.RetryBackoff(attempts))
					}
					output, err := t.client.BatchWriteItem(&dynamodb.BatchWriteItemInput{
						RequestItems: unprocessed,
					})
					if err != nil {
						errc <- err
					}
					unprocessed = output.UnprocessedItems
					attempts++
				}
			}(req)
			req = []*dynamodb.WriteRequest{}
		}
	}
	go func() {
		wg.Wait()
		close(errc)
	}()
	return <-errc
}

func (t *Truncator) truncate(table string) error {
	meta, err := t.readMeta(table)
	if err != nil {
		return err
	}
	keys := []*string{}
	keySchema := meta.Table.KeySchema
	for _, k := range keySchema {
		keys = append(keys, k.AttributeName)
	}
	tableSize := *meta.Table.TableSizeBytes
	totalSegments := int64(math.Ceil(float64(tableSize) / megabyte))
	totalSegments = calc.Min(totalSegments, maxTotalSegments)
	if totalSegments == 0 {
		cfmt.Warningf("Table '%s' has no items.\n", table)
		return nil
	}

	// Delete all keys
	cfmt.Successf("[%d/%d] Truncating the table '%s'...\n", 0, totalSegments, table)
	errc := make(chan error, 1)
	wg := sync.WaitGroup{}
	wg.Add(int(totalSegments))
	for i := int64(0); i < totalSegments; i++ {
		go func(segment int64) {
			defer wg.Done()
			cfmt.Infof("[%d/%d] Deleting the %d segment of table '%s'...\n", segment+1, totalSegments, segment, table)
			startKey := map[string]*dynamodb.AttributeValue{}
			attempts := 0
			for {
				if attempts > 0 {
					time.Sleep(retryer.RetryBackoff(attempts))
				}
				scanned, err := t.client.Scan(&dynamodb.ScanInput{
					TableName:         aws.String(table),
					AttributesToGet:   keys,
					ExclusiveStartKey: startKey,
					Segment:           aws.Int64(segment),
					TotalSegments:     aws.Int64(totalSegments),
				})
				if err != nil {
					errc <- err
				}
				if err = t.delete(table, scanned); err != nil {
					errc <- err
				}
				startKey = scanned.LastEvaluatedKey
				if len(startKey) == 0 {
					break
				}
			}
			cfmt.Successf("[%d/%d] The %d segment of table '%s' was deleted.\n", segment+1, totalSegments, segment, table)
		}(i)
	}
	go func() {
		wg.Wait()
		cfmt.Successf("[%d/%d] Table '%s' was truncated successfully.\n", totalSegments, totalSegments, table)
		close(errc)
	}()
	return <-errc
}

func (t *Truncator) recreate(table string) error {
	meta, err := t.readMeta(table)
	if err != nil {
		return err
	}

	// Delete the table and wait until complete
	cfmt.Infof("Deleting the table '%s'...\n", table)
	t.client.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String(table),
	})
	err = t.client.WaitUntilTableNotExists(&dynamodb.DescribeTableInput{
		TableName: aws.String(table),
	})
	if err != nil {
		return err
	}
	cfmt.Successf("Table '%s' was deleted.\n", table)

	// Make create table input
	cfmt.Infof("Recreating the table '%s'...\n", table)
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: meta.Table.AttributeDefinitions,
		KeySchema:            meta.Table.KeySchema,
		TableName:            meta.Table.TableName,
	}
	if meta.Table.BillingModeSummary != nil {
		input.SetBillingMode(*meta.Table.BillingModeSummary.BillingMode)
	} else {
		input.SetProvisionedThroughput(&dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  meta.Table.ProvisionedThroughput.ReadCapacityUnits,
			WriteCapacityUnits: meta.Table.ProvisionedThroughput.WriteCapacityUnits,
		})
	}
	if meta.Table.StreamSpecification != nil {
		input.SetStreamSpecification(meta.Table.StreamSpecification)
	}
	globalSecondaryIndexes := []*dynamodb.GlobalSecondaryIndex{}
	for _, v := range meta.Table.GlobalSecondaryIndexes {
		globalSecondaryIndexes = append(globalSecondaryIndexes, &dynamodb.GlobalSecondaryIndex{
			IndexName:  v.IndexName,
			KeySchema:  v.KeySchema,
			Projection: v.Projection,
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  v.ProvisionedThroughput.ReadCapacityUnits,
				WriteCapacityUnits: v.ProvisionedThroughput.WriteCapacityUnits,
			},
		})
	}
	if len(globalSecondaryIndexes) > 0 {
		input.SetGlobalSecondaryIndexes(globalSecondaryIndexes)
	}
	localSecondaryIndexes := []*dynamodb.LocalSecondaryIndex{}
	for _, v := range meta.Table.LocalSecondaryIndexes {
		localSecondaryIndexes = append(localSecondaryIndexes, &dynamodb.LocalSecondaryIndex{
			IndexName:  v.IndexName,
			KeySchema:  v.KeySchema,
			Projection: v.Projection,
		})
	}
	if len(globalSecondaryIndexes) > 0 {
		input.SetLocalSecondaryIndexes(localSecondaryIndexes)
	}

	// Create the table and wait until complete
	_, err = t.client.CreateTable(input)
	if err != nil {
		return err
	}
	err = t.client.WaitUntilTableExists(&dynamodb.DescribeTableInput{
		TableName: aws.String(table),
	})
	if err != nil {
		return err
	}
	cfmt.Successf("Table '%s' was recreated successfully.\n", table)
	return nil
}

// Truncate truncates the dynamodb tables
func (t *Truncator) Truncate(tables []string, willRecreate bool) []error {
	errs := make([]error, 0)
	wg := sync.WaitGroup{}
	for _, table := range tables {
		wg.Add(1)
		go func(table string) {
			defer wg.Done()
			var err error
			if willRecreate {
				err = t.recreate(table)
			} else {
				err = t.truncate(table)
			}
			if err != nil {
				errs = append(errs, err)
			}
		}(table)
	}
	wg.Wait()
	return errs
}
