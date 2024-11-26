// Copyright 2022 The go-rajchain Authors
// This file is part of the go-rajchain library.
//
// The go-rajchain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-rajchain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-rajchain library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/rajchain/go-rajchain/beacon/blsync"
	"github.com/rajchain/go-rajchain/cmd/utils"
	"github.com/rajchain/go-rajchain/internal/debug"
	"github.com/rajchain/go-rajchain/internal/flags"
	"github.com/rajchain/go-rajchain/log"
	"github.com/rajchain/go-rajchain/node"
	"github.com/rajchain/go-rajchain/rpc"
	"github.com/urfave/cli/v2"
)

func main() {
	app := flags.NewApp("beacon light syncer tool")
	app.Flags = slices.Concat([]cli.Flag{
		utils.BeaconApiFlag,
		utils.BeaconApiHeaderFlag,
		utils.BeaconThresholdFlag,
		utils.BeaconNoFilterFlag,
		utils.BeaconConfigFlag,
		utils.BeaconGenesisRootFlag,
		utils.BeaconGenesisTimeFlag,
		utils.BeaconCheckpointFlag,
		//TODO datadir for optional permanent database
		utils.MainnetFlag,
		utils.SepoliaFlag,
		utils.HoleskyFlag,
		utils.BlsyncApiFlag,
		utils.BlsyncJWTSecretFlag,
	},
		debug.Flags,
	)
	app.Before = func(ctx *cli.Context) error {
		flags.MigrateGlobalFlags(ctx)
		return debug.Setup(ctx)
	}
	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		return nil
	}
	app.Action = sync

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func sync(ctx *cli.Context) error {
	// set up blsync
	client := blsync.NewClient(utils.MakeBeaconLightConfig(ctx))
	client.SetEngineRPC(makeRPCClient(ctx))
	client.Start()

	// run until stopped
	<-ctx.Done()
	client.Stop()
	return nil
}

func makeRPCClient(ctx *cli.Context) *rpc.Client {
	if !ctx.IsSet(utils.BlsyncApiFlag.Name) {
		log.Warn("No engine API target specified, performing a dry run")
		return nil
	}
	if !ctx.IsSet(utils.BlsyncJWTSecretFlag.Name) {
		utils.Fatalf("JWT secret parameter missing") //TODO use default if datadir is specified
	}

	engineApiUrl, jwtFileName := ctx.String(utils.BlsyncApiFlag.Name), ctx.String(utils.BlsyncJWTSecretFlag.Name)
	var jwtSecret [32]byte
	if jwt, err := node.ObtainJWTSecret(jwtFileName); err == nil {
		copy(jwtSecret[:], jwt)
	} else {
		utils.Fatalf("Error loading or generating JWT secret: %v", err)
	}
	auth := node.NewJWTAuth(jwtSecret)
	cl, err := rpc.DialOptions(context.Background(), engineApiUrl, rpc.WithHTTPAuth(auth))
	if err != nil {
		utils.Fatalf("Could not create RPC client: %v", err)
	}
	return cl
}