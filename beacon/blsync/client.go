// Copyright 2024 The go-rajchain Authors
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

package blsync

import (
	"github.com/rajchain/go-rajchain/beacon/light"
	"github.com/rajchain/go-rajchain/beacon/light/api"
	"github.com/rajchain/go-rajchain/beacon/light/request"
	"github.com/rajchain/go-rajchain/beacon/light/sync"
	"github.com/rajchain/go-rajchain/beacon/params"
	"github.com/rajchain/go-rajchain/beacon/types"
	"github.com/rajchain/go-rajchain/common/mclock"
	"github.com/rajchain/go-rajchain/ethdb/memorydb"
	"github.com/rajchain/go-rajchain/event"
	"github.com/rajchain/go-rajchain/rpc"
)

type Client struct {
	urls         []string
	customHeader map[string]string
	config       *params.ClientConfig
	scheduler    *request.Scheduler
	blockSync    *beaconBlockSync
	engineRPC    *rpc.Client

	chainHeadSub event.Subscription
	engineClient *engineClient
}

func NewClient(config params.ClientConfig) *Client {
	// create data structures
	var (
		db             = memorydb.New()
		committeeChain = light.NewCommitteeChain(db, &config.ChainConfig, config.Threshold, !config.NoFilter)
		headTracker    = light.NewHeadTracker(committeeChain, config.Threshold)
	)
	headSync := sync.NewHeadSync(headTracker, committeeChain)

	// set up scheduler and sync modules
	scheduler := request.NewScheduler()
	checkpointInit := sync.NewCheckpointInit(committeeChain, config.Checkpoint)
	forwardSync := sync.NewForwardUpdateSync(committeeChain)
	beaconBlockSync := newBeaconBlockSync(headTracker)
	scheduler.RegisterTarget(headTracker)
	scheduler.RegisterTarget(committeeChain)
	scheduler.RegisterModule(checkpointInit, "checkpointInit")
	scheduler.RegisterModule(forwardSync, "forwardSync")
	scheduler.RegisterModule(headSync, "headSync")
	scheduler.RegisterModule(beaconBlockSync, "beaconBlockSync")

	return &Client{
		scheduler:    scheduler,
		urls:         config.Apis,
		customHeader: config.CustomHeader,
		config:       &config,
		blockSync:    beaconBlockSync,
	}
}

func (c *Client) SetEngineRPC(engine *rpc.Client) {
	c.engineRPC = engine
}

func (c *Client) Start() error {
	headCh := make(chan types.ChainHeadEvent, 16)
	c.chainHeadSub = c.blockSync.SubscribeChainHead(headCh)
	c.engineClient = startEngineClient(c.config, c.engineRPC, headCh)

	c.scheduler.Start()
	for _, url := range c.urls {
		beaconApi := api.NewBeaconLightApi(url, c.customHeader)
		c.scheduler.RegisterServer(request.NewServer(api.NewApiServer(beaconApi), &mclock.System{}))
	}
	return nil
}

func (c *Client) Stop() error {
	c.engineClient.stop()
	c.chainHeadSub.Unsubscribe()
	c.scheduler.Stop()
	return nil
}
