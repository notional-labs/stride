package keeper_test

import (
	testkeeper "github.com/Stride-Labs/stride/v4/testutil/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	"testing"
)

func TestQueryModuleAddress(t *testing.T) {
	keeper, ctx := testkeeper.StakeibcKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	Name := "ModuleAccountTest"

	add := keeper.SetTestModuleAccount(ctx, Name)

	response, err := keeper.ModuleAddress(wctx, &types.QueryModuleAddressRequest{Name: Name})

	require.NoError(t, err)
	require.Equal(t, &types.QueryModuleAddressResponse{Addr: add.String()}, response)
}
