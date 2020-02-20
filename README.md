# AWS Graphql DynamoDB Pager

This lambda function implements a pager for a dynamodb datasource. It allows
graphql standard compliant paging without the sparse results that would otherwise
be returned by the default dynamodb paging behaviour.

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

## LICENSE

Copyright (c) 2020 Crown Copyright (Office for National Statistics)

Released under MIT license, see [LICENSE](LICENSE) for details.
