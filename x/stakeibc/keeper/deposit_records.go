package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v4/utils"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (k Keeper) CreateDepositRecordsForEpoch(ctx sdk.Context, epochNumber uint64) {
	k.Logger(ctx).Info(fmt.Sprintf("Creating Deposit Records for Epoch %d", epochNumber))

	for _, hostZone := range k.GetAllHostZone(ctx) {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Creating Deposit Record"))

		depositRecord := recordstypes.DepositRecord{
			Amount:             sdk.ZeroInt(),
			Denom:              hostZone.HostDenom,
			HostZoneId:         hostZone.ChainId,
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: epochNumber,
		}
		k.RecordsKeeper.AppendDepositRecord(ctx, depositRecord)
	}
}
func (k Keeper) TransferExistingDepositsToHostZones(ctx sdk.Context, epochNumber uint64, depositRecords []recordstypes.DepositRecord) {
	transferDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		isTransferRecord := record.Status == recordstypes.DepositRecord_TRANSFER_QUEUE
		isBeforeCurrentEpoch := record.DepositEpochNumber < epochNumber
		return isTransferRecord && isBeforeCurrentEpoch
	})

	ibcTransferTimeoutNanos := k.GetParam(ctx, types.KeyIBCTransferTimeoutNanos)

	for _, depositRecord := range transferDepositRecords {
		k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId,
			"Processing deposit record %d: %v%s", depositRecord.Id, depositRecord.Amount, depositRecord.Denom))

		// if a TRANSFER_QUEUE record has 0 balance and was created in the previous epoch, it's safe to remove since it will never be updated or used
		if depositRecord.Amount.LTE(sdk.ZeroInt()) && depositRecord.DepositEpochNumber < epochNumber {
			k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId, "Empty deposit record - Removing."))
			k.RecordsKeeper.RemoveDepositRecord(ctx, depositRecord.Id)
			continue
		}

		hostZone, hostZoneFound := k.GetHostZone(ctx, depositRecord.HostZoneId)
		if !hostZoneFound {
			k.Logger(ctx).Error(fmt.Sprintf("[TransferExistingDepositsToHostZones] Host zone not found for deposit record id %d", depositRecord.Id))
			continue
		}

		hostZoneModuleAddress := hostZone.GetAddress()
		delegateAccount := hostZone.GetDelegationAccount()
		if delegateAccount == nil || delegateAccount.GetAddress() == "" {
			k.Logger(ctx).Error(fmt.Sprintf("[TransferExistingDepositsToHostZones] Zone %s is missing a delegation address!", hostZone.ChainId))
			continue
		}
		delegateAddress := delegateAccount.GetAddress()

		k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId, "Transferring %v%s", depositRecord.Amount, hostZone.HostDenom))
		transferCoin := sdk.NewCoin(hostZone.IbcDenom, depositRecord.Amount)

		// timeout 30 min in the future
		// NOTE: this assumes no clock drift between chains, which tendermint guarantees
		// if we onboard non-tendermint chains, we need to use the time on the host chain to
		// calculate the timeout
		// https://github.com/tendermint/tendermint/blob/v0.34.x/spec/consensus/bft-time.md
		timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + ibcTransferTimeoutNanos
		msg := ibctypes.NewMsgTransfer(ibctypes.PortID, hostZone.TransferChannelId, transferCoin, hostZoneModuleAddress, delegateAddress, clienttypes.Height{}, timeoutTimestamp)
		k.Logger(ctx).Info(fmt.Sprintf("TransferExistingDepositsToHostZones msg %v", msg))

		// transfer the deposit record and update its status to TRANSFER_IN_PROGRESS
		err := k.RecordsKeeper.Transfer(ctx, msg, depositRecord)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("\t[TransferExistingDepositsToHostZones] Failed to initiate IBC transfer to host zone, HostZone: %v, Channel: %v, Amount: %v, ModuleAddress: %v, DelegateAddress: %v, Timeout: %v",
				hostZone.ChainId, hostZone.TransferChannelId, transferCoin, hostZoneModuleAddress, delegateAddress, timeoutTimestamp))
			k.Logger(ctx).Error(fmt.Sprintf("\t[TransferExistingDepositsToHostZones] err {%s}", err.Error()))
			continue
		}
	}
}

func (k Keeper) StakeExistingDepositsOnHostZones(ctx sdk.Context, epochNumber uint64, depositRecords []recordstypes.DepositRecord) {
	stakeDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		isStakeRecord := record.Status == recordstypes.DepositRecord_DELEGATION_QUEUE
		isBeforeCurrentEpoch := record.DepositEpochNumber < epochNumber
		return isStakeRecord && isBeforeCurrentEpoch
	})

	// limit the number of staking deposits to process per epoch
	maxDepositRecordsToStake := utils.Min(len(stakeDepositRecords), cast.ToInt(k.GetParam(ctx, types.KeyMaxStakeICACallsPerEpoch)))
	k.Logger(ctx).Info(fmt.Sprintf("Staking %d out of %d deposit records", maxDepositRecordsToStake, len(stakeDepositRecords)))

	for _, depositRecord := range stakeDepositRecords[:maxDepositRecordsToStake] {
		k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId,
			"Processing deposit record %d: %v%s", depositRecord.Id, depositRecord.Amount, depositRecord.Denom))

		hostZone, hostZoneFound := k.GetHostZone(ctx, depositRecord.HostZoneId)
		if !hostZoneFound {
			k.Logger(ctx).Error(fmt.Sprintf("[StakeExistingDepositsOnHostZones] Host zone not found for deposit record {%d}", depositRecord.Id))
			continue
		}

		delegateAccount := hostZone.GetDelegationAccount()
		if delegateAccount == nil || delegateAccount.GetAddress() == "" {
			k.Logger(ctx).Error(fmt.Sprintf("[StakeExistingDepositsOnHostZones] Zone %s is missing a delegation address!", hostZone.ChainId))
			continue
		}

		k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId, "Staking %v%s", depositRecord.Amount, hostZone.HostDenom))
		stakeAmount := sdk.NewCoin(hostZone.HostDenom, depositRecord.Amount)

		err := k.DelegateOnHost(ctx, hostZone, stakeAmount, depositRecord)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Did not stake %s on %s | err: %s", stakeAmount.String(), hostZone.ChainId, err.Error()))
			continue
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("Successfully submitted stake for %s on %s", stakeAmount.String(), hostZone.ChainId))
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("hostZone", hostZone.ChainId),
				sdk.NewAttribute("newAmountStaked", depositRecord.Amount.String()),
			),
		)
	}
}
