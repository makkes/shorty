package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// DynamoDB encapsulates all DynamoDB-related database functions
type DynamoDB struct {
	svc       *dynamodb.DynamoDB
	tableName string
}

type ShortenedURL struct {
	Key string `json:"key"`
	Url string `json:"url"`
}

// NewDynamoDB creates and initializes a new DynamoDB connection
func NewDynamoDB() (*DynamoDB, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("eu-west-1"),
		Credentials: credentials.NewSharedCredentials("", "default"),
	})
	if err != nil {
		return nil, err
	}

	svc := dynamodb.New(sess)

	_, err = svc.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("key"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("key"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String("shorty"),
	})

	if err != nil {
		// when a ResourceInUseException occurs we assume the table already exists and continue
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() != dynamodb.ErrCodeResourceInUseException {
			return nil, err
		}
	}

	return &DynamoDB{
		svc:       svc,
		tableName: "shorty",
	}, nil
}

func (db *DynamoDB) GetURL(key []byte) ([]byte, error) {
	res, err := db.svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"key": {
				S: aws.String(string(key)),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	shortenedUrl := ShortenedURL{}
	err = dynamodbattribute.UnmarshalMap(res.Item, &shortenedUrl)
	if err != nil {
		return nil, err
	}

	if shortenedUrl.Key == "" {
		return nil, nil
	}

	// TODO: increase view counter for item

	return []byte(shortenedUrl.Url), nil
}

func (db *DynamoDB) SaveURL(url string, keybuffer <-chan []byte) (string, error) {
	key := string(<-keybuffer)
	shortenedUrl := ShortenedURL{
		Key: key,
		Url: url,
	}

	av, err := dynamodbattribute.MarshalMap(shortenedUrl)
	if err != nil {
		return "", err
	}

	input := &dynamodb.PutItemInput{
		Item:                av,
		TableName:           aws.String(db.tableName),
		ConditionExpression: aws.String("attribute_not_exists(Key)"),
	}

	_, err = db.svc.PutItem(input)
	if err != nil {
		return "", err
	}
	return key, nil
}
