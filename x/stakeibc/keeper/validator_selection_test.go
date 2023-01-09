package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	"fmt"

	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone_InvalidAmountOfDelegation() {
	tc := s.SetupGetHostZoneUnbondingMsgs()

	unbond := sdk.ZeroInt()
	_, err := s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, tc.hostZone, unbond)
	s.Require().EqualError(err, stakeibc.ErrNoValidatorWeights.Error(), "Delegate zero amount should fail")

}

func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone_ErrNoValidatorsWeight() {
	tc := s.SetupGetHostZoneUnbondingMsgs()

	unbond := sdk.NewIntFromUint64(1_000_000)

	// assign zero amount to all validators's weights
	validators := tc.hostZone.GetValidators()
	for _, validator := range validators {
		validator.Weight = 0
	}

	// if weight is zero then return err
	_, err := s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, tc.hostZone, unbond)
	s.Require().EqualError(err, stakeibc.ErrNoValidatorWeights.Error(), "Delegate zero amount should fail")
}

func (s *KeeperTestSuite) SetupGetValidatorDelegationAmtDifferences(validators []*stakeibc.Validator) stakeibc.HostZone {
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)
	delegationAddr := "cosmos_DELEGATION"

	delegationAccount := stakeibc.ICAAccount{
		Address: delegationAddr,
		Target:  stakeibc.ICAAccountType_DELEGATION,
	}

	hostZone := stakeibc.HostZone{
		ChainId:           "GAIA",
		HostDenom:         "uatom",
		Bech32Prefix:      "cosmos",
		Validators:        validators,
		DelegationAccount: &delegationAccount,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	return hostZone
}

func (s *KeeperTestSuite) TestGetValidatorDelegationAmtDifferences_Successful() {
	validators := []*stakeibc.Validator{
		{
			Address:       "cosmos_VALIDATOR",
			DelegationAmt: sdk.NewIntFromUint64(1_000_000),
			Weight:        uint64(1),
		},
	}
	hostZone := s.SetupGetValidatorDelegationAmtDifferences(validators)

	_, err := s.App.StakeibcKeeper.GetValidatorDelegationAmtDifferences(s.Ctx, hostZone)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestGetValidatorDelegationAmtDifferences_ErrorGetTargetValAtmsForHostZone() {
	validators := []*stakeibc.Validator{
		{
			Address:       "cosmos_VALIDATOR",
			DelegationAmt: sdk.NewIntFromUint64(0),
			Weight:        uint64(2),
		},
	}
	hostZone := s.SetupGetValidatorDelegationAmtDifferences(validators)
	_, err := s.App.StakeibcKeeper.GetValidatorDelegationAmtDifferences(s.Ctx, hostZone)
	s.Require().Error(err)
	s.Require().Equal(err, types.ErrNoValidatorWeights)
}
