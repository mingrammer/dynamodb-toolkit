package mock

import (
	"sync"
	"time"
	"unsafe"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type table struct {
	desc  *dynamodb.TableDescription
	items []map[string]*dynamodb.AttributeValue
}

// DynamoDBClient is mocking the dynamodb
type DynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	tables map[string]*table
	mutex  *sync.Mutex // For concurrent request
}

// NewDynamoDBClient creates a mocked dynamodb client
func NewDynamoDBClient() *DynamoDBClient {
	return &DynamoDBClient{
		tables: map[string]*table{},
		mutex:  new(sync.Mutex),
	}
}

// BatchWriteItem mocks the dynamodb BatchWriteItem operation
func (d *DynamoDBClient) BatchWriteItem(input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
	d.mutex.Lock()
	for t, ri := range input.RequestItems {
		if _, ok := d.tables[t]; !ok {
			return nil, awserr.New(dynamodb.ErrCodeResourceNotFoundException, "Not found", nil)
		}
		for _, r := range ri {
			if r.DeleteRequest != nil {
				for i, it := range d.tables[t].items {
					match := true
					for a, k := range r.DeleteRequest.Key {
						if it[a] != k {
							match = false
							break
						}
					}
					if match {
						d.tables[t].items = append(d.tables[t].items[:i], d.tables[t].items[i+1:]...)
						*d.tables[t].desc.ItemCount--
						*d.tables[t].desc.TableSizeBytes -= int64(unsafe.Sizeof(it)) // It is not actual size in bytes
					}
				}
			}
			if r.PutRequest != nil {
				d.tables[t].items = append(d.tables[t].items, r.PutRequest.Item)
				*d.tables[t].desc.ItemCount++
				*d.tables[t].desc.TableSizeBytes += int64(unsafe.Sizeof(r.PutRequest.Item)) // It is not actual size in bytes
			}
		}
	}
	d.mutex.Unlock()
	return &dynamodb.BatchWriteItemOutput{}, nil
}

// CreateTable is mocking the dynamodb CreateTable operation
func (d *DynamoDBClient) CreateTable(input *dynamodb.CreateTableInput) (*dynamodb.CreateTableOutput, error) {
	name := *input.TableName
	table := &table{
		desc: &dynamodb.TableDescription{
			AttributeDefinitions: input.AttributeDefinitions,
			CreationDateTime:     aws.Time(time.Now()),
			ItemCount:            aws.Int64(0),
			KeySchema:            input.KeySchema,
			TableName:            &name,
			TableSizeBytes:       aws.Int64(0),
		},
		items: []map[string]*dynamodb.AttributeValue{},
	}
	if input.BillingMode != nil {
		table.desc.SetBillingModeSummary(&dynamodb.BillingModeSummary{
			BillingMode: input.BillingMode,
		})
	} else {
		table.desc.SetProvisionedThroughput(&dynamodb.ProvisionedThroughputDescription{
			ReadCapacityUnits:  input.ProvisionedThroughput.ReadCapacityUnits,
			WriteCapacityUnits: input.ProvisionedThroughput.WriteCapacityUnits,
		})
	}
	d.tables[name] = table
	return &dynamodb.CreateTableOutput{
		TableDescription: table.desc,
	}, nil
}

// DeleteTable is mocking the dynamodb DeleteTable operation
func (d *DynamoDBClient) DeleteTable(input *dynamodb.DeleteTableInput) (*dynamodb.DeleteTableOutput, error) {
	name := *input.TableName
	desc := d.tables[name].desc
	delete(d.tables, name)
	return &dynamodb.DeleteTableOutput{
		TableDescription: desc,
	}, nil
}

// DescribeTable is mocking the dynamodb DescribeTable operation
func (d *DynamoDBClient) DescribeTable(input *dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error) {
	name := *input.TableName
	if table, ok := d.tables[name]; ok {
		return &dynamodb.DescribeTableOutput{
			Table: table.desc,
		}, nil
	}
	return nil, awserr.New(dynamodb.ErrCodeResourceNotFoundException, "Not found", nil)
}

// Scan is mocking the dynamodb Scan operation
func (d *DynamoDBClient) Scan(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	name := *input.TableName
	if _, ok := d.tables[name]; !ok {
		return nil, awserr.New(dynamodb.ErrCodeResourceNotFoundException, "Not found", nil)
	}
	total := int64(len(d.tables[name].items))
	l := int(*input.Segment * total / *input.TotalSegments)
	r := int((*input.Segment + 1) * total / *input.TotalSegments)
	items := []map[string]*dynamodb.AttributeValue{}
	for _, it := range d.tables[name].items[l:r] {
		attr := map[string]*dynamodb.AttributeValue{}
		for _, a := range input.AttributesToGet {
			attr[*a] = it[*a]
		}
		items = append(items, attr)
	}
	return &dynamodb.ScanOutput{
		Count:        aws.Int64(int64(len(items))),
		Items:        items,
		ScannedCount: aws.Int64(int64(len(items))),
	}, nil
}

// WaitUntilTableExists is mocking the dynamodb WaitUntilTableExists operation
func (d *DynamoDBClient) WaitUntilTableExists(input *dynamodb.DescribeTableInput) error {
	name := *input.TableName
	if _, ok := d.tables[name]; !ok {
		return awserr.New(dynamodb.ErrCodeResourceNotFoundException, "Table exists", nil)
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

// WaitUntilTableNotExists is mocking the dynamodb WaitUntilTableNotExists operation
func (d *DynamoDBClient) WaitUntilTableNotExists(input *dynamodb.DescribeTableInput) error {
	name := *input.TableName
	if _, ok := d.tables[name]; ok {
		return awserr.New(dynamodb.ErrCodeResourceNotFoundException, "Table not exists", nil)
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}
