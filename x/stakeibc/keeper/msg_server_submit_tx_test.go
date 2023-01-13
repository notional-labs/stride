package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	_ "github.com/stretchr/testify/suite"
)

type DelegateOnHostTestcase struct {
	hostZone      stakeibctypes.HostZone
	amt           sdk.Coin
	depositRecord []recordstypes.DepositRecord
}

func (s *KeeperTestSuite) SetupSetWithdrawalAddressOnHost_emptyStrideEpoch() DelegateOnHostTestcase {
	//Set Deposit Records of type Delegate
	stakedBal := sdk.NewInt(5_000)
	AddressDelegateAccount := "1"
	DelegateAccount := &types.ICAAccount{Address: AddressDelegateAccount, Target: types.ICAAccountType_DELEGATION}
	AddressWithdrawalAccount := "stride_ADDRESS"
	WithdrawalAccount := &types.ICAAccount{Address: AddressWithdrawalAccount, Target: types.ICAAccountType_WITHDRAWAL}
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	withdrawalAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "WITHDRAWAL")
	s.CreateICAChannel(delegationAccountOwner)
	s.CreateICAChannel(withdrawalAccountOwner)
	DepositRecordDelegate := []recordstypes.DepositRecord{
		{
			Id:         1,
			Amount:     sdk.NewInt(1000),
			Denom:      Atom,
			HostZoneId: HostChainId,
			Status:     recordstypes.DepositRecord_DELEGATION_QUEUE,
		},
		{
			Id:         2,
			Amount:     sdk.NewInt(3000),
			Denom:      Atom,
			HostZoneId: HostChainId,
			Status:     recordstypes.DepositRecord_DELEGATION_QUEUE,
		},
	}
	stakeAmount := sdk.NewInt(1_000_000)
	stakeCoin := sdk.NewCoin(Atom, stakeAmount)

	//  define the host zone with stakedBal and validators with staked amounts
	hostVal1Addr := "cosmos_VALIDATOR_1"
	hostVal2Addr := "cosmos_VALIDATOR_2"
	amtVal1 := sdk.NewInt(1_000_000)
	amtVal2 := sdk.NewInt(2_000_000)
	wgtVal1 := uint64(1)
	wgtVal2 := uint64(2)

	validators := []*stakeibctypes.Validator{
		{
			Address:       hostVal1Addr,
			DelegationAmt: amtVal1,
			Weight:        wgtVal1,
		},
		{
			Address:       hostVal2Addr,
			DelegationAmt: amtVal2,
			Weight:        wgtVal2,
		},
	}

	hostZone := stakeibctypes.HostZone{
		ChainId:           HostChainId,
		HostDenom:         Atom,
		IbcDenom:          IbcAtom,
		RedemptionRate:    sdk.NewDec(1.0),
		StakedBal:         stakedBal,
		Validators:        validators,
		DelegationAccount: DelegateAccount,
		WithdrawalAccount: WithdrawalAccount,
		ConnectionId:      ibctesting.FirstConnectionID,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return DelegateOnHostTestcase{
		hostZone:      hostZone,
		amt:           stakeCoin,
		depositRecord: DepositRecordDelegate,
	}
}

func (s *KeeperTestSuite) SetupSetWithdrawalAddressOnHost_valid() DelegateOnHostTestcase {
	tc := s.SetupSetWithdrawalAddressOnHost_emptyStrideEpoch()
	//set Stride epoch
	epochIdentifier := epochtypes.STRIDE_EPOCH
	epochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochIdentifier,
		EpochNumber:        uint64(2),
		NextEpochStartTime: uint64(time.Now().UnixNano()),
		Duration:           uint64(2),
	}

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	return DelegateOnHostTestcase{
		hostZone:      tc.hostZone,
		amt:           tc.amt,
		depositRecord: tc.depositRecord,
	}
}

func (s *KeeperTestSuite) TestSetWithdrawalAddressOnHost_successful() {
	tc := s.SetupSetWithdrawalAddressOnHost_valid()
	err := s.App.StakeibcKeeper.SetWithdrawalAddressOnHost(s.Ctx, tc.hostZone)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestSetWithdrawalAddressOnHost_FailedToGetICATimeoutNanos() {
	tc := s.SetupSetWithdrawalAddressOnHost_emptyStrideEpoch()

	err := s.App.StakeibcKeeper.SetWithdrawalAddressOnHost(s.Ctx, tc.hostZone)
	expectedErr := "Failed to SubmitTxs for connection-0, GAIA, [delegator_address:\"1\" withdraw_address:\"stride_ADDRESS\" ]: "
	expectedErr += "invalid request"
	s.EqualError(err, expectedErr, "Hostzone is set without Stride Epoch so it should fail")
}
