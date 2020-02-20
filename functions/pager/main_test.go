package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/matryer/is"
)

func TestEncodeKey(t *testing.T) {

	is := is.New(t)

	key := []*dynamodb.KeySchemaElement{
		{
			AttributeName: aws.String("id"),
			KeyType:       aws.String("HASH"),
		},
	}

	n := node{
		"id":      "11111",
		"fluffy":  true,
		"species": "panda",
	}

	expectedCursor := "eyJpZCI6eyJCIjpudWxsLCJCT09MIjpudWxsLCJCUyI6bnVsbCwiTCI6bnVsbCwiTSI6bnVsbCwiTiI6bnVsbCwiTlMiOm51bGwsIk5VTEwiOm51bGwsIlMiOiIxMTExMSIsIlNTIjpudWxsfX0="

	gotCursor, err := encodeKey(key, n)
	is.NoErr(err)
	is.Equal(gotCursor, expectedCursor)
}

func TestDecodeKey(t *testing.T) {

	is := is.New(t)

	encoded := "eyJpZCI6eyJCIjpudWxsLCJCT09MIjpudWxsLCJCUyI6bnVsbCwiTCI6bnVsbCwiTSI6bnVsbCwiTiI6bnVsbCwiTlMiOm51bGwsIk5VTEwiOm51bGwsIlMiOiIxMTExMSIsIlNTIjpudWxsfX0="

	expected := map[string]*dynamodb.AttributeValue{
		"id": &dynamodb.AttributeValue{
			S: aws.String("11111"),
		},
	}

	decoded, err := decodeKey(encoded)
	is.NoErr(err)
	is.Equal(decoded, expected)
}
