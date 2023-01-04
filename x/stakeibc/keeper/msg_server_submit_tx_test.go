package keeper_test

import (
	"fmt"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestGetStartTimeNextEpoch_Success() {
	// SetEpochTracker
	epochIdentifier := epochtypes.STRIDE_EPOCH
	epochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochIdentifier,
		EpochNumber:        uint64(2),
		NextEpochStartTime: uint64(time.Now().Unix()),
		Duration:           uint64(2),
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	epochReturn, err := s.App.StakeibcKeeper.GetStartTimeNextEpoch(s.Ctx, epochIdentifier)

	s.Require().NoError(err)

	s.Require().Equal(epochReturn, epochTracker.NextEpochStartTime)

}
func (s *KeeperTestSuite) TestGetStartTimeNextEpoch_FailedToGetEpoch() {
	// SetEpochTracker
	epochIdentifier := epochtypes.STRIDE_EPOCH
	epochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochIdentifier,
		EpochNumber:        uint64(2),
		NextEpochStartTime: uint64(time.Now().Unix()),
		Duration:           uint64(2),
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	_, err := s.App.StakeibcKeeper.GetStartTimeNextEpoch(s.Ctx, "epoch_stride")

	s.Require().EqualError(err, fmt.Sprintf("Failed to get epoch tracker for %s: %s", "epoch_stride", sdkerrors.ErrInvalidRequest.Error()))
}
