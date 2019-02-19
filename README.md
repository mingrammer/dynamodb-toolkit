<br><br>

<h1 align="center">DynamoDB Toolkit</h1>

<p align="center">
  <a href="/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg"/></a>
  <a href="https://godoc.org/github.com/mingrammer/dynamodb-toolkit"><img src="https://godoc.org/github.com/mingrammer/dynamodb-toolkit?status.svg"/></a>
  <a href="https://goreportcard.com/report/github.com/mingrammer/dynamodb-toolkit"><img src="https://goreportcard.com/badge/github.com/mingrammer/dynamodb-toolkit"/></a>
  <a href="https://travis-ci.org/mingrammer/dynamodb-toolkit"><img src="https://travis-ci.org/mingrammer/dynamodb-toolkit.svg?branch=master"/></a>
</p>


<p align="center">
A command line dynamodb toolkit
</p>

<br><br><br>

dynamotk is a command line tool for aws dynamodb that provides some useful utilities that are not supported by dynamodb operations and official aws cli tools. 

## Installation

### Using go get

> Go version 1.10 or higher is required.

```
go get github.com/mingrammer/dynamodb-toolkit
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
- coming soon... (maybe dump/restore features)

## Usage

### Truncate

```bash
# Truncate the `user` table from local dynamodb.
$ dynamotk --endpoint http://localhost:8000 truncate --table-names user

# Truncate the `user`, `item` tables from aws dynamodb with default credentials.
$ dynamotk truncate --table-names user,item

# You can also pass the `access key id`, `secret access key`, `profile` and `region` optionally. (see `dynamotk -h`)
$ dynamotk --access-key-id xxx --secret-access-key xxx --region ap-northeast-2 truncate --table-names user,item

# Truncation is just (concurrently) repeating the delete operations for all keys. So if your tables are big, there could be cost overhead.
# In this case, you can use `--recreate` option. It will not use delete operations, but just delete the table itself and recreate the table while preserving the table description.
$ dynamotk --profile prod truncate --table-names largetable --recreate
```

## License

MIT