// Copyright 2021 The go-rajchain Authors
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

// Package ethconfig contains the configuration of the ETH and LES protocols.
package ethconfig

import (
	"fmt"
	"time"

	"github.com/rajchain/go-rajchain/common"
	"github.com/rajchain/go-rajchain/consensus"
	"github.com/rajchain/go-rajchain/consensus/beacon"
	"github.com/rajchain/go-rajchain/consensus/clique"
	"github.com/rajchain/go-rajchain/consensus/ethash"
	"github.com/rajchain/go-rajchain/core"
	"github.com/rajchain/go-rajchain/core/txpool/blobpool"
	"github.com/rajchain/go-rajchain/core/txpool/legacypool"
	"github.com/rajchain/go-rajchain/eth/downloader"
	"github.com/rajchain/go-rajchain/eth/gasprice"
	"github.com/rajchain/go-rajchain/ethdb"
	"github.com/rajchain/go-rajchain/log"
	"github.com/rajchain/go-rajchain/miner"
	"github.com/rajchain/go-rajchain/params"
)

// FullNodeGPO contains default gasprice oracle settings for full node.
var FullNodeGPO = gasprice.Config{
	Blocks:           20,
	Percentile:       60,
	MaxHeaderHistory: 1024,
	MaxBlockHistory:  1024,
	MaxPrice:         gasprice.DefaultMaxPrice,
	IgnorePrice:      gasprice.DefaultIgnorePrice,
}

// Defaults contains default settings for use on the rajchain main net.
var Defaults = Config{
	SyncMode:           downloader.SnapSync,
	NetworkId:          0, // enable auto configuration of networkID == chainID
	TxLookupLimit:      2350000,
	TransactionHistory: 2350000,
	StateHistory:       params.FullImmutabilityThreshold,
	DatabaseCache:      512,
	TrieCleanCache:     154,
	TrieDirtyCache:     256,
	TrieTimeout:        60 * time.Minute,
	SnapshotCache:      102,
	FilterLogCacheSize: 32,
	Miner:              miner.DefaultConfig,
	TxPool:             legacypool.DefaultConfig,
	BlobPool:           blobpool.DefaultConfig,
	RPCGasCap:          50000000,
	RPCEVMTimeout:      5 * time.Second,
	GPO:                FullNodeGPO,
	RPCTxFeeCap:        1, // 1 ether
}

//go:generate go run github.com/fjl/gencodec -type Config -formats toml -out gen_config.go

// Config contains configuration options for ETH and LES protocols.
type Config struct {
	// The genesis block, which is inserted if the database is empty.
	// If nil, the rajchain main net block is used.
	Genesis *core.Genesis `toml:",omitempty"`

	// Network ID separates blockchains on the peer-to-peer networking level. When left
	// zero, the chain ID is used as network ID.
	NetworkId uint64
	SyncMode  downloader.SyncMode

	// This can be set to list of enrtree:// URLs which will be queried for
	// nodes to connect to.
	EthDiscoveryURLs  []string
	SnapDiscoveryURLs []string

	// State options.
	NoPruning  bool // Whether to disable pruning and flush everything to disk
	NoPrefetch bool // Whether to disable prefetching and only load state on demand

	// Deprecated: use 'TransactionHistory' instead.
	TxLookupLimit uint64 `toml:",omitempty"` // The maximum number of blocks from head whose tx indices are reserved.

	TransactionHistory uint64 `toml:",omitempty"` // The maximum number of blocks from head whose tx indices are reserved.
	StateHistory       uint64 `toml:",omitempty"` // The maximum number of blocks from head whose state histories are reserved.

	// State scheme represents the scheme used to store rajchain states and trie
	// nodes on top. It can be 'hash', 'path', or none which means use the scheme
	// consistent with persistent state.
	StateScheme string `toml:",omitempty"`

	// RequiredBlocks is a set of block number -> hash mappings which must be in the
	// canonical chain of all remote peers. Setting the option makes geth verify the
	// presence of these blocks for every new peer connection.
	RequiredBlocks map[uint64]common.Hash `toml:"-"`

	// Database options
	SkipBcVersionCheck bool `toml:"-"`
	DatabaseHandles    int  `toml:"-"`
	DatabaseCache      int
	DatabaseFreezer    string

	TrieCleanCache int
	TrieDirtyCache int
	TrieTimeout    time.Duration
	SnapshotCache  int
	Preimages      bool

	// This is the number of blocks for which logs will be cached in the filter system.
	FilterLogCacheSize int

	// Mining options
	Miner miner.Config

	// Transaction pool options
	TxPool   legacypool.Config
	BlobPool blobpool.Config

	// Gas Price Oracle options
	GPO gasprice.Config

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool

	// Enables VM tracing
	VMTrace           string
	VMTraceJsonConfig string

	// RPCGasCap is the global gas cap for eth-call variants.
	RPCGasCap uint64

	// RPCEVMTimeout is the global timeout for eth-call.
	RPCEVMTimeout time.Duration

	// RPCTxFeeCap is the global transaction fee(price * gaslimit) cap for
	// send-transaction variants. The unit is ether.
	RPCTxFeeCap float64

	// OverrideCancun (TODO: remove after the fork)
	OverrideCancun *uint64 `toml:",omitempty"`

	// OverrideVerkle (TODO: remove after the fork)
	OverrideVerkle *uint64 `toml:",omitempty"`
}

// CreateConsensusEngine creates a consensus engine for the given chain config.
// Clique is allowed for now to live standalone, but ethash is forbidden and can
// only exist on already merged networks.
func CreateConsensusEngine(config *params.ChainConfig, db ethdb.Database) (consensus.Engine, error) {
	if config.TerminalTotalDifficulty == nil {
		log.Error("Geth only supports PoS networks. Please transition legacy networks using Geth v1.13.x.")
		return nil, fmt.Errorf("'terminalTotalDifficulty' is not set in genesis block")
	}
	// Wrap previously supported consensus engines into their post-merge counterpart
	if config.Clique != nil {
		return beacon.New(clique.New(config.Clique, db)), nil
	}
	return beacon.New(ethash.NewFaker()), nil
}
