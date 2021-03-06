<br><br>

<h1 align="center">DynamoDB Toolkit</h1>

<p align="center">
  <a href="/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg"/></a>
  <a href="https://godoc.org/github.com/mingrammer/dynamodb-toolkit"><img src="https://godoc.org/github.com/mingrammer/dynamodb-toolkit?status.svg"/></a>
  <a href="https://goreportcard.com/report/github.com/mingrammer/dynamodb-toolkit"><img src="https://goreportcard.com/badge/github.com/mingrammer/dynamodb-toolkit"/></a>
  <a href="https://travis-ci.org/mingrammer/dynamodb-toolkit"><img src="https://travis-ci.org/mingrammer/dynamodb-toolkit.svg?branch=master"/></a>
</p>


<p align="center">
A command line toolkit for aws dynamodb
</p>

<br><br><br>

dynamotk is a command line toolkit for aws dynamodb that provides some useful utilities that are not supported by dynamodb operations and official aws cli tools.

## Installation

### Using go get

> Go version 1.13 or higher is required.

```
go get github.com/mingrammer/dynamodb-toolkit/cmd/dynamotk
```

### Using [homebrew](https://brew.sh)

```
brew tap mingrammer/homebrew-taps
brew install dynamodb-toolkit
```

### Using .tar.gz archive

Download gzip file from [Github Releases](https://github.com/mingrammer/dynamodb-toolkit/releases/latest) according to your OS. Then, copy the unzipped executable to under system path.

## Features

- Table truncate
- Coming soon... (maybe dump/restore features)

## Usage

### Truncate

```console
# Truncate the `user` table from local dynamodb.
dynamotk --endpoint http://localhost:8000 truncate --table-names user

# Truncate the `user`, `item` tables from aws dynamodb with default credentials.
dynamotk truncate --table-names user,item

# You can also pass the `access key id`, `secret access key`, `profile` and `region` optionally. (see `dynamotk -h`)
dynamotk --access-key-id xxx --secret-access-key xxx truncate --table-names user,item

# Truncation is just (concurrently) repeating the delete operations for all keys.
# So if your tables are big, it can cause cost overhead.
# In this case, you can use `--recreate` option.
# It will delete the table itself and recreate the table while preserving the description.
dynamotk --profile prod --region ap-northeast-2 truncate --table-names largetable --recreate
```

## Known issues

When throttling happens, `dynamotk` does not retry read or write (delete request), so some items could be remaining not deleted. I should support `backoff-retry` algorithm to fix it.

For now, you should run the `truncate` command multiple times until the table becomes empty to overcome this issue or use `--recreate` option.

## License

MIT
