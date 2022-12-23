package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	types_mint "github.com/Stride-Labs/stride/v4/x/mint/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (k Keeper) GetTestModuleAccount(ctx sdk.Context, Name string) authtypes.AccountI {
	acctAddress := k.GetsubModuleAddress(Name)
	acc := k.accountKeeper.GetAccount(ctx, acctAddress)
	return acc
}

func (k Keeper) SetTestModuleAccount(ctx sdk.Context, Name string) sdk.Address {
	acctAddress := k.GetsubModuleAddress(Name)

	acc := k.accountKeeper.NewAccount(
		ctx,
		authtypes.NewModuleAccount(
			authtypes.NewBaseAccountWithAddress(acctAddress),
			acctAddress.String(),
		),
	)
	k.Logger(ctx).Info(fmt.Sprintf("Created new %s.%s module account %s", types.ModuleName, Name, acc.GetAddress().String()))

	k.accountKeeper.SetAccount(ctx, acc)

	return acctAddress
}

func (k Keeper) GetsubModuleAddress(Name string) sdk.AccAddress {
	key := []byte(Name)
	return address.Module(types_mint.ModuleName, key)
}
