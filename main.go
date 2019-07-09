package main

import (
	"encoding/json"
	"flag"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"go.uber.org/zap"
)

type Table struct {
	Name    string   `json:"TableName"`
	ADefs   []AttDef `json:"AttributeDefinitions"`
	KSchema []KeyDef `json:"KeySchema"`
	LSI     []Index
	GI      []Index
}

type AttDef struct {
	Name  string `json:"AttributeName"`
	AType string `json:"AttributeType"`
}

type KeyDef struct {
	Name  string `json:"AttributeName"`
	KType string `json:"KeyType"`
}

type Index struct {
	Name       string
	Projection string
	KSchema    []KeyDef
}

func main() {
	pl, _ := zap.NewProduction()
	l := pl.Sugar()

	sess, err := session.NewSession()
	if err != nil {
		l.Info("Could not create AWS session...exiting")
		os.Exit(1)
	}

	unsafe := flag.Bool("unsafe", false, "allows running without DB_ENDPOINT and DB_TABLE_PREFIX")
	flag.Parse()

	endpoint := os.Getenv("DB_ENDPOINT")
	tablePrefix := os.Getenv("DB_TABLE_PREFIX")
	if (len(tablePrefix) == 0 || len(endpoint) == 0) && !*unsafe {
		l.Info("must provide DB_TABLE_PREFIX and DB_ENDPOINT")
		os.Exit(1)
	}

	tbls := []Table{}
	for _, v := range os.Environ() {
		envar := strings.Split(v, "=")
		if strings.Contains(envar[0], "TABLE_DEFINITION_") {
			t := &Table{}
			err := json.Unmarshal([]byte(envar[1]), t)
			if err != nil {
				l.Infow("error unmarshalling table defintion", "envar", envar[0], "error", err.Error())
				os.Exit(1)
			}
			tbls = append(tbls, *t)
		}
	}

	client := dynamodb.New(sess, &aws.Config{Endpoint: aws.String(endpoint)})

	for _, t := range tbls {
		l.Infow("creating table", "name", t.Name)
		dbI := &dynamodb.CreateTableInput{
			TableName: aws.String(tablePrefix + t.Name),
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(1),
				WriteCapacityUnits: aws.Int64(1)}}

		for _, ad := range t.ADefs {
			dbI.AttributeDefinitions = append(
				dbI.AttributeDefinitions,
				&dynamodb.AttributeDefinition{AttributeName: aws.String(ad.Name), AttributeType: aws.String(ad.AType)})
		}

		for _, ks := range t.KSchema {
			dbI.KeySchema = append(
				dbI.KeySchema,
				&dynamodb.KeySchemaElement{AttributeName: aws.String(ks.Name), KeyType: aws.String(ks.KType)})
		}

		for _, li := range t.LSI {
			lin := &dynamodb.LocalSecondaryIndex{
				IndexName:  aws.String(li.Name),
				Projection: &dynamodb.Projection{ProjectionType: aws.String(li.Projection)}}

			for _, ks := range li.KSchema {
				lin.KeySchema = append(
					lin.KeySchema,
					&dynamodb.KeySchemaElement{AttributeName: aws.String(ks.Name), KeyType: aws.String(ks.KType)})
			}

			dbI.LocalSecondaryIndexes = append(dbI.LocalSecondaryIndexes, lin)
		}

		for _, gi := range t.GI {
			gin := &dynamodb.GlobalSecondaryIndex{
				IndexName: aws.String(gi.Name),
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(1),
					WriteCapacityUnits: aws.Int64(1)},
				Projection: &dynamodb.Projection{ProjectionType: aws.String(gi.Projection)}}

			for _, ks := range gi.KSchema {
				gin.KeySchema = append(
					gin.KeySchema,
					&dynamodb.KeySchemaElement{AttributeName: aws.String(ks.Name), KeyType: aws.String(ks.KType)})
			}

			dbI.GlobalSecondaryIndexes = append(dbI.GlobalSecondaryIndexes, gin)
		}

		_, err := client.CreateTable(dbI)

		if err != nil {
			awse := err.(awserr.Error)
			if awse.Code() != "ResourceInUseException" {
				l.Infow("unable to create table", "name", t.Name, "input", dbI, "error", awse)
				os.Exit(1)
			}
			l.Info("table already exists")
		}
	}

	l.Info("finished")
}
