// Copyright 2023 The go-rajchain Authors
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

package sync

import (
	"github.com/rajchain/go-rajchain/beacon/light/request"
	"github.com/rajchain/go-rajchain/beacon/types"
	"github.com/rajchain/go-rajchain/common"
)

var (
	EvNewHead             = &request.EventType{Name: "newHead"}             // data: types.HeadInfo
	EvNewOptimisticUpdate = &request.EventType{Name: "newOptimisticUpdate"} // data: types.OptimisticUpdate
	EvNewFinalityUpdate   = &request.EventType{Name: "newFinalityUpdate"}   // data: types.FinalityUpdate
)

type (
	ReqUpdates struct {
		FirstPeriod, Count uint64
	}
	RespUpdates struct {
		Updates    []*types.LightClientUpdate
		Committees []*types.SerializedSyncCommittee
	}
	ReqHeader  common.Hash
	RespHeader struct {
		Header               types.Header
		Canonical, Finalized bool
	}
	ReqCheckpointData common.Hash
	ReqBeaconBlock    common.Hash
	ReqFinality       struct{}
)
