package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

func (k Keeper) ModuleAddress(goCtx context.Context, req *types.QueryModuleAddressRequest) (*types.QueryModuleAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	Name := req.Name
	acctAddress := sdk.AccAddress(address.Module(types.ModuleName, []byte(Name)))
	addr := k.accountKeeper.GetAccount(ctx, acctAddress).GetAddress().String()

	return &types.QueryModuleAddressResponse{Addr: addr}, nil
}
