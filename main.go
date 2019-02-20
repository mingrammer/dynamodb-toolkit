package main

import (
	"errors"
	"os"
	"strings"

	"github.com/mingrammer/cfmt"
	"github.com/mingrammer/dynamodb-toolkit/config"
	"github.com/mingrammer/dynamodb-toolkit/service"
	"github.com/mingrammer/dynamodb-toolkit/toolkit"
	"github.com/urfave/cli"
)

// CLI information
const (
	name      = "dynamotk"
	author    = "mingrammer"
	email     = "mingrammer@gmail.com"
	version   = "0.0.2"
	usage     = "A command line toolkit for AWS DyanmoDB"
	usageText = "dynamotk [OPTIONS] command [OPTIONS/FLAGS]"
)

func main() {
	app := cli.NewApp()
	app.Name = name
	app.Author = author
	app.Email = email
	app.Version = version
	app.Usage = usage
	app.UsageText = usageText
	app.Flags = buildGlobalFlags()
	app.Before = buildBeforeFunc()
	app.Commands = []cli.Command{
		buildTruncateCommand(),
	}
	err := app.Run(os.Args)
	if err != nil {
		cfmt.Errorln(err.Error())
	}
}

func buildBeforeFunc() cli.BeforeFunc {
	return func(ctx *cli.Context) error {
		config.SetCredentials(
			ctx.String("access-key-id"),
			ctx.String("secret-access-key"),
		)
		config.SetProfile(ctx.String("profile"))
		config.SetRegion(ctx.String("region"))
		config.SetEndpoint(ctx.String("endpoint"))
		return nil
	}
}

func buildGlobalFlags() []cli.Flag {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:   "access-key-id",
			Usage:  "aws access key id",
			EnvVar: "AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-access-key",
			Usage:  "aws secret access key",
			EnvVar: "AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "profile",
			Usage:  "aws credential profile",
			EnvVar: "AWS_PROFILE",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "dynamodb region",
			EnvVar: "AWS_REGION",
		},
		cli.StringFlag{
			Name:   "endpoint",
			Usage:  "dynamodb endpoint. It is for local dynamodb",
			EnvVar: "AWS_DYNAMODB_ENDPOINT",
		},
	}
	return flags
}

func buildTruncateCommand() cli.Command {
	cmd := cli.Command{
		Name:  "truncate",
		Usage: "truncate the dynamodb tables",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "table-names",
				Usage: "comma delimited table names which will be truncated",
			},
			cli.BoolFlag{
				Name:  "recreate",
				Usage: "delete and recreate the tables. It is useful for large tables",
			},
		},
		Action: func(ctx *cli.Context) error {
			tablesString := ctx.String("table-names")
			if len(tablesString) == 0 {
				return errors.New(cfmt.Serror("You must pass at least one table name"))
			}
			tables := strings.Split(tablesString, ",")
			client, err := service.NewDynamoDBClient()
			if err != nil {
				return err
			}
			truncator := toolkit.NewTruncator(client)
			willRecreate := ctx.Bool("recreate")
			if errs := truncator.Truncate(tables, willRecreate); len(errs) > 0 {
				for _, err := range errs {
					cfmt.Errorln(err.Error())
				}
			}
			return nil
		},
	}
	return cmd
}
