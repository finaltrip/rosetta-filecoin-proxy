// +build rosetta_rpc

package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	"math"
)

const FactorSecondToMillisecond int64 = 1e3

type SyncStatus struct {
	currentHeight   []int64
	globalSyncState api.SyncStateStage
}

func (status SyncStatus) IsSynced() bool {

	if status.globalSyncState == api.StageSyncComplete {
		return true
	}

	return false
}

func (status SyncStatus) GetMaxHeight() int64 {
	var max int64
	for _, height := range status.currentHeight {
		if height > max {
			max = height
		}
	}

	if max > 0 {
		max--
	}

	return max
}

func (status SyncStatus) GetMinHeight() int64 {

	if len(status.currentHeight) == 0 {
		return 0
	}

	var min int64 = math.MaxInt64
	for _, height := range status.currentHeight {
		if height < min {
			min = height
		}
	}

	return min
}

func CheckSyncStatus(ctx context.Context, node *api.FullNode) (*SyncStatus, *types.Error) {

	fullAPI := *node
	syncState, err := fullAPI.SyncState(ctx)

	if err != nil || len(syncState.ActiveSyncs) == 0 {
		return nil, ErrUnableToGetSyncStatus
	}

	var (
		status = SyncStatus{
			globalSyncState: api.StageIdle,
		}
		syncComplete = false
		earliestStat = api.StageIdle
	)

	for _, w := range syncState.ActiveSyncs {
		if w.Target == nil {
			continue
		}

		switch w.Stage {
		case api.StageSyncErrored:
			return nil, ErrSyncErrored
		case api.StageSyncComplete:
			syncComplete = true
		default:
			if w.Stage > earliestStat {
				earliestStat = w.Stage
			}
		}

		status.currentHeight = append(status.currentHeight, int64(w.Height))
	}

	if syncComplete {
		status.globalSyncState = api.StageSyncComplete
	} else {
		status.globalSyncState = earliestStat
	}

	return &status, nil
}
