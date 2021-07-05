// Copyright 2021 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mendersoftware/go-lib-micro/config"
	"github.com/urfave/cli"

	"github.com/mendersoftware/reporting/app/indexer"
	"github.com/mendersoftware/reporting/app/server"
	dconfig "github.com/mendersoftware/reporting/config"
	"github.com/mendersoftware/reporting/store"
)

func main() {
	doMain(os.Args)
}

func doMain(args []string) {
	var configPath string

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "Configuration `FILE`. Supports JSON, TOML, YAML and HCL formatted configs.",
				Value:       "config.yaml",
				Destination: &configPath,
			},
		},
		Commands: []cli.Command{
			{
				Name:   "server",
				Usage:  "Run the HTTP API server",
				Action: cmdServer,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "automigrate",
						Usage: "Run database migrations before starting.",
					},
				},
			},
			{
				Name:   "indexer",
				Usage:  "Run the indexer process",
				Action: cmdIndexer,
				Flags: []cli.Flag{
					&cli.Int64Flag{
						Name:  "devices",
						Usage: "Number of devices to index",
						Value: 1000,
					},
					&cli.StringFlag{
						Name:  "tenant_id",
						Usage: "Destination tenant ID",
						Value: "test-tenant",
					},
					&cli.BoolFlag{
						Name:  "automigrate",
						Usage: "Run database migrations before starting.",
					},
				},
			},
			{
				Name:   "migrate",
				Usage:  "Run the migrations",
				Action: cmdMigrate,
			},
		},
	}
	app.Usage = "Reporting"
	app.Version = "1.0.0"
	app.Action = cmdServer

	app.Before = func(args *cli.Context) error {
		err := config.FromConfigFile(configPath, dconfig.Defaults)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("error loading configuration: %s", err),
				1)
		}

		// Enable setting config values by environment variables
		config.Config.SetEnvPrefix("REPORTING")
		config.Config.AutomaticEnv()
		config.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

		return nil
	}

	err := app.Run(args)
	if err != nil {
		log.Fatal(err)
	}
}

func cmdServer(args *cli.Context) error {
	store, err := getStore(args)
	if err != nil {
		return err
	}
	if args.Bool("automigrate") {
		ctx := context.Background()
		err := store.Migrate(ctx)
		if err != nil {
			return err
		}
	}
	return server.InitAndRun(config.Config, store)
}

func cmdIndexer(args *cli.Context) error {
	store, err := getStore(args)
	if err != nil {
		return err
	}
	if args.Bool("automigrate") {
		ctx := context.Background()
		err := store.Migrate(ctx)
		if err != nil {
			return err
		}
	}
	devices := args.Int64("devices")
	tid := args.String("tenant_id")
	return indexer.InitAndRun(config.Config, store, devices, tid)
}

func cmdMigrate(args *cli.Context) error {
	store, err := getStore(args)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return store.Migrate(ctx)
}

func getStore(args *cli.Context) (store.Store, error) {
	addresses := config.Config.GetStringSlice(dconfig.SettingElasticsearchAddresses)
	store, err := store.NewStore(
		store.WithServerAddresses(addresses),
	)
	if err != nil {
		return nil, err
	}
	return store, nil
}
