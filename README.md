# AWS Graphql DynamoDB Pager

This lambda function implements a pager for a dynamodb datasource. It allows
graphql standard compliant paging without the sparse results that would otherwise
be returned by the default dynamodb paging behaviour.

- [AWS Graphql DynamoDB Pager](#aws-graphql-dynamodb-pager)
  - [Permissions](#permissions)
  - [Build](#build)
  - [Graphql Resolver](#graphql-resolver)
    - [Request](#request)
    - [Response](#response)
  - [Example Query](#example-query)
  - [LICENSE](#license)

## Permissions

This repo does not contain IAM policies/roles. To deploy you will need to create
suitable policies to access your datasources.

As a minium you'll need the following permissions on the table(s) you wish to
access:

- `dynamodb:Scan`
- `dyanmodb:DescribeTable`

This lambda does not alter any data so does not need any write/update/delete permissions.

For writing _CloudWatch_ logs you'll also require on the appropriate resources:

- `logs:CreateLogGroup`
- `logs:CreateLogStream`
- `logs:PutLogEvents`

## Build

The build is standard as you would expect for a Go AWS lambda:

```shell
# Assume already in the repo
cd functions/pager
GOOS=linux go build main.go
zip function.zip main
```

You can now deploy the compiled lambda to AWS as usual, selecting runtime type
as `Go 1.x`

## Graphql Resolver

The following gives suitable templates to use for resolvers calling this lambda

### Request

Need to supply `<TABLE_NAME>`

```text
{
  "version" : "2017-02-28",
  "operation": "Invoke",
  "payload": {
    "tableName": "<TABLE_NAME>",
    "filter": $util.transform.toDynamoDBFilterExpression($ctx.args.filter),
    #if($context.args.first) "first": $util.toJson($context.args.first), #end
    #if($context.args.after) "after": $util.toJson($context.args.after), #end
  }
}
```

### Response

No need to JSONify the result as it's passed back correctly encoded by the lambda

```text
$context.result
```

## Example Query

Query:

```graphql
query list($first: Int, $after: String){
  listAnimals(
    first: $first,
    after: $after,
    filter: {
    species:{
      contains: "panda"
    }
  }){
    pageInfo{
      hasNextPage
      endCursor
    }
    edges{
      cursor
      node {
        id
        fluffy
        species
      }
    }
  }
}
```

Response:

```json
{
  "data": {
    "listAnimals": {
      "pageInfo": {
        "hasNextPage": true,
        "endCursor": "eyJpZCI6eyJCIjpudWxsLCJCT09MIjpudWxsLCJCUyI6bnVsbCwiTCI6bnVsbCwiTSI6bnVsbCwiTiI6bnVsbCwiTlMiOm51bGwsIk5VTEwiOm51bGwsIlMiOiIxMTExMSIsIlNTIjpudWxsfX0="
      },
      "edges": [
        {
          "cursor": "eyJpZCI6eyJCIjpudWxsLCJCT09MIjpudWxsLCJCUyI6bnVsbCwiTCI6bnVsbCwiTSI6bnVsbCwiTiI6bnVsbCwiTlMiOm51bGwsIk5VTEwiOm51bGwsIlMiOiIxMTExMSIsIlNTIjpudWxsfX0=",
          "node": {
            "id": "11111",
            "fluffy": true,
            "species": "panda"
          }
        }
      ]
    }
  }
}
```

## LICENSE

Copyright (c) 2020 Crown Copyright (Office for National Statistics)

Released under MIT license, see [LICENSE](LICENSE) for details.
