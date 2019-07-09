builddb
---

Takes the json from the `aws dynamodb describe-table` output and uses it to rebuild the tables. Useful for running in CI/CD scripts.

It looks in the environment for any varible that start with `TABLE_DEFINITION_`  and expects that to be a json description of the table.

For Example:

```json
{
  "TableName": "test_table",
  "AttributeDefinitions": [
    {
      "AttributeName": "cluster",
      "AttributeType": "S"
    },
    {
      "AttributeName": "channel",
      "AttributeType": "S"
    }
  ],
  "KeySchema": [
    {
      "AttributeName": "cluster",
      "KeyType": "HASH"
    },
    {
      "AttributeName": "channel",
      "KeyType": "RANGE"
    }
  ]
```

would create a table with the name `test_table` and a hash key of `cluster` and range key of `channel`