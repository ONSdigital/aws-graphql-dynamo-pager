package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"
)

const (
	defaultPageSize int64 = 100
)

type (
	payload struct {
		TableName string                  `json:"tableName"`
		Filter    *dynamoFilterExpression `json:"filter"`

		// Read limits
		First int64 `json:"first"`
		Last  int64 `json:"last"` // TODO not supported yet

		// Cursors
		After  string `json:"after"`
		Before string `json:"before"` // TOOD not supported yet
	}

	dynamoFilterExpression struct {
		Expression string                              `json:"expression"`
		Names      map[string]*string                  `json:"expressionNames"`
		Values     map[string]*dynamodb.AttributeValue `json:"expressionValues"`
	}

	// Pagination sections: https://facebook.github.io/relay/graphql/connections.htm
	pageInfo struct {
		HasNext     bool   `json:"hasNextPage"`
		HasPrev     bool   `json:"hasPreviousPage"`
		StartCursor string `json:"startCursor"`
		EndCursor   string `json:"endCursor"`
	}

	edge struct {
		Cursor string `json:"cursor"`
		Node   *node  `json:"node"`
	}

	node map[string]interface{}

	response struct {
		Edges []edge `json:"edges"`
		// Nodes    []node   `json:"nodes"` // Not supporting direct nodes yet
		PageInfo pageInfo `json:"pageInfo"`
		// Look into supporting totalCount
	}
)

// HandleRequest handles the incoming request to the lambda
func HandleRequest(ctx context.Context, p payload) (string, error) {

	if p.TableName == "" {
		return "", errors.New("request missing tableName")
	}

	if p.Filter == nil {
		return "", errors.New("request missing filter expression")
	}

	if p.Before != "" || p.Last > 0 {
		return "", errors.New("backwards paging not yet supported")
	}

	// Check for pagination items being passed in and ensure we don't have
	// invalid combinations!
	// TODO
	// ...

	if p.First == 0 {
		p.First = defaultPageSize
	}

	// TODO Look into whether to create the session for each call or to create
	// once and share
	svc := dynamodb.New(session.New())

	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  p.Filter.Names,
		ExpressionAttributeValues: p.Filter.Values,
		FilterExpression:          aws.String(p.Filter.Expression),
		TableName:                 aws.String(p.TableName),
		Limit:                     aws.Int64(p.First),
	}
	if p.After != "" {
		after, err := decodeKey(p.After)
		if err != nil {
			return "", errors.Wrap(err, "unable to decode 'after' cursor")
		}
		input.ExclusiveStartKey = after
	}

	// Determine the keyfields for the table so we can construct the next/prev
	// cursors (especially when we end not strictly on the LastKey)
	table, err := svc.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: &p.TableName,
	})
	if err != nil {
		return "", err
	}

	tableKeySchema := table.Table.KeySchema
	if len(tableKeySchema) == 0 {
		return "", errors.New("empty table key schema - unable to determine key fields")
	}

	// Pre-size the results set to be at most the upper limit. We
	// may actually return less if there are fewer matching items
	resp := response{
		PageInfo: pageInfo{},
		Edges:    make([]edge, 0, p.First),
	}

	// Need to keep reading data until any of:
	// - No more data (empty LastEvaluatedKey)
	// - Read enough items to fulfil the given `limit`
	var found int64 = 0
	for int64(len(resp.Edges)) < p.First {

		result, err := svc.ScanWithContext(ctx, input)
		if err != nil {
			return "", err
		}

		for _, item := range result.Items {
			found++

			e := edge{
				Cursor: "",
				Node:   nil,
			}

			var n node
			err := dynamodbattribute.UnmarshalMap(item, &n)
			if err != nil {
				return "", err
			}
			e.Node = &n

			cursor, err := encodeKey(tableKeySchema, n)
			if err != nil {
				return "", errors.Wrap(err, "failed to encode key")
			}
			e.Cursor = cursor

			resp.Edges = append(resp.Edges, e)

			if found == p.First {
				log.Printf("read '%d' items, want '%d', returning", found, p.First)
				// If we've found as many as the client asked for, we cut and return
				// what we have - no point reading further into what we don't need!
				// In this case we return the cursor of the last item
				resp.PageInfo.HasNext = true
				resp.PageInfo.EndCursor = cursor
				break
			}
		}

		if result.LastEvaluatedKey == nil {
			// No more data to be returned at all
			log.Printf("read '%d' items, want '%d', no more data to return", len(resp.Edges), p.First)
			break
		}

		// To read the next chunk, we need to update the scan
		// input to include the last key
		log.Printf("read '%d' items, want '%d', reading next page", len(resp.Edges), p.First)
		input.ExclusiveStartKey = result.LastEvaluatedKey
	}

	// TODO - deal with HAS PREVIOUS and PREVIOUS KEY
	// ...
	// ...

	output, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func encodeKey(tableKeySchema []*dynamodb.KeySchemaElement, n node) (string, error) {
	key := make(map[string]*dynamodb.AttributeValue)
	for _, element := range tableKeySchema {
		v, err := dynamodbattribute.Marshal(n[*element.AttributeName])
		if err != nil {
			return "", err
		}
		key[*element.AttributeName] = v
	}

	marshaled, err := json.Marshal(key)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(marshaled), err
}

func decodeKey(encoded string) (map[string]*dynamodb.AttributeValue, error) {
	jsn, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var key map[string]*dynamodb.AttributeValue
	err = json.Unmarshal(jsn, &key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func main() {
	lambda.Start(HandleRequest)
}
