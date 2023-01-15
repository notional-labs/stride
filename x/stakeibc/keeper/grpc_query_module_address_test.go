package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	"github.com/cosmos/cosmos-sdk/types/address"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (s *KeeperTestSuite) SetupTestQueryModuleAddress() (sdk.Address, string) {
	Name := "stakeibc"
	ak := s.App.AccountKeeper

	acctAddress := sdk.AccAddress(address.Module(types.ModuleName, []byte(Name)))
	acc := ak.NewAccount(
		s.Ctx,
		authtypes.NewModuleAccount(
			authtypes.NewBaseAccountWithAddress(acctAddress),
			acctAddress.String(),
		),
	)
	ak.SetAccount(s.Ctx, acc)
	return acctAddress, Name
}
func (s *KeeperTestSuite) TestQueryModuleAddress_Successful() {
	add, Name := s.SetupTestQueryModuleAddress()
	wctx := sdk.WrapSDKContext(s.Ctx)

	response, err := s.App.StakeibcKeeper.ModuleAddress(wctx, &types.QueryModuleAddressRequest{Name: Name})
	s.Require().NoError(err)
	s.Require().Equal(&types.QueryModuleAddressResponse{Addr: add.String()}, response)
}
func (s *KeeperTestSuite) TestQueryModuleAddress_InvalidRequest() {
	wctx := sdk.WrapSDKContext(s.Ctx)

	_, err := s.App.StakeibcKeeper.ModuleAddress(wctx, nil)
	s.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "invalid request"))
}
