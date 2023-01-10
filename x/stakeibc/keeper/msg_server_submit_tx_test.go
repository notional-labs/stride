package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	ibctesting "github.com/cosmos/ibc-go/v3/testing"
)

type SubmitTxTestCase struct {
	epochTracker stakeibctypes.EpochTracker
	hostZone     stakeibctypes.HostZone
	channel      string
	msgs         []sdk.Msg
}

func (s *KeeperTestSuite) SetupSubMitTxs() SubmitTxTestCase {
	stakedBal := sdk.NewInt(1_000_000)
	AddressWithdrawalAccount := "2"

	WithdrawalAccount := &types.ICAAccount{Address: AddressWithdrawalAccount, Target: types.ICAAccountType_WITHDRAWAL}

	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	channelID := s.CreateICAChannel(delegationAccountOwner)

	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	hostZone := stakeibctypes.HostZone{
		ChainId:           HostChainId,
		HostDenom:         Atom,
		IbcDenom:          IbcAtom,
		RedemptionRate:    sdk.NewDec(1.0),
		StakedBal:         stakedBal,
		WithdrawalAccount: WithdrawalAccount,
		DelegationAccount: &stakeibctypes.ICAAccount{Address: delegationAddress},
		ConnectionId:      ibctesting.FirstConnectionID,
	}

	delegationIca := hostZone.DelegationAccount
	withdrawalIcaAddr := hostZone.WithdrawalAccount.Address
	msgs := []sdk.Msg{
		&distributiontypes.MsgSetWithdrawAddress{
			DelegatorAddress: delegationIca.Address,
			WithdrawAddress:  withdrawalIcaAddr,
		},
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	epochIdentifier := epochtypes.STRIDE_EPOCH
	epochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochIdentifier,
		EpochNumber:        uint64(2),
		NextEpochStartTime: uint64(time.Now().UnixNano()),
		Duration:           uint64(2),
	}

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)
	return SubmitTxTestCase{
		epochTracker: epochTracker,
		hostZone:     hostZone,
		channel:      channelID,
		msgs:         msgs,
	}
}
func (s *KeeperTestSuite) TestSubmitsTx_Successful() {
	tc := s.SetupSubMitTxs()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount

	_, err := s.App.StakeibcKeeper.SubmitTxs(s.Ctx, hostZone.ConnectionId, msgs, *account, tc.epochTracker.NextEpochStartTime, "", nil)
	s.Require().NoError(err)

}
func (s *KeeperTestSuite) TestSubmitsTx_FailedCallBackSendTx() {
	tc := s.SetupSubMitTxs()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount
	//test case
	tc.epochTracker.NextEpochStartTime = 0

	_, err := s.App.StakeibcKeeper.SubmitTxs(s.Ctx, hostZone.ConnectionId, msgs, *account, tc.epochTracker.NextEpochStartTime, "", nil)
	s.Require().EqualError(err, "timeout timestamp must be in the future")
}

func (s *KeeperTestSuite) TestSubmitsTx_FailedToRetrieveActiveChannel() {
	tc := s.SetupSubMitTxs()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount
	//test case
	account.Target = stakeibctypes.ICAAccountType_FEE

	_, err := s.App.StakeibcKeeper.SubmitTxs(s.Ctx, hostZone.ConnectionId, msgs, *account, tc.epochTracker.NextEpochStartTime, "", nil)
	s.Require().EqualError(err, "failed to retrieve active channel for port icacontroller-GAIA.FEE: no active channel for this owner")

}

func (s *KeeperTestSuite) TestSubmitsTx_InvalidConnectionIdNotFound() {
	tc := s.SetupSubMitTxs()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount
	//test case
	ConnectionID_test := "connection_test"

	_, err := s.App.StakeibcKeeper.SubmitTxs(s.Ctx, ConnectionID_test, msgs, *account, tc.epochTracker.NextEpochStartTime, "", nil)
	s.Require().EqualError(err, fmt.Sprintf("invalid connection id, %s not found", ConnectionID_test))
}
func (s *KeeperTestSuite) TestSubmitTxsEpoch_Successful() {
	tc := s.SetupSubMitTxs()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount
	connectionID := ibctesting.FirstConnectionID
	epochIdentifier := epochtypes.STRIDE_EPOCH

	_, err := s.App.StakeibcKeeper.SubmitTxsEpoch(s.Ctx, connectionID, msgs, *account, epochIdentifier, "", nil)
	s.Require().NoError(err)
}
func (s *KeeperTestSuite) TestSubmitTxsEpoch_FailedCallBackSubmitTxs() {
	tc := s.SetupSubMitTxs()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount
	//test case
	connectionID := "connectionID_test"
	epochIdentifier := epochtypes.STRIDE_EPOCH

	_, err := s.App.StakeibcKeeper.SubmitTxsEpoch(s.Ctx, connectionID, msgs, *account, epochIdentifier, "", nil)
	s.Require().EqualError(err, fmt.Sprintf("invalid connection id, %s not found", connectionID))
}
func (s *KeeperTestSuite) TestSubmitTxsEpoch_FailedToGetICA() {
	tc := s.SetupSubMitTxs()
	hostZone := tc.hostZone
	msgs := tc.msgs
	account := hostZone.DelegationAccount
	connectionID := ibctesting.FirstConnectionID
	//test case
	epochIdentifier := "epochtypes_test"

	_, err := s.App.StakeibcKeeper.SubmitTxsEpoch(s.Ctx, connectionID, msgs, *account, epochIdentifier, "", nil)
	s.Require().EqualError(err, fmt.Sprintf("Failed to convert timeoutNanos to uint64, error: Failed to get epoch tracker for %s: invalid request: invalid request", epochIdentifier))
}
