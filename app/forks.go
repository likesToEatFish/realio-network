package app

import (
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var (
	ForkHeight        = int64(5989487)
	oneEnternityLater = time.Date(9999, 9, 9, 9, 9, 9, 9, time.UTC)
)

// ScheduleForkUpgrade executes any necessary fork logic for based upon the current
// block height and chain ID (mainnet or testnet). It sets an upgrade plan once
// the chain reaches the pre-defined upgrade height.
//
// CONTRACT: for this logic to work properly it is required to:
//
//  1. Release a non-breaking patch version so that the chain can set the scheduled upgrade plan at upgrade-height.
//  2. Release the software defined in the upgrade-info
func (app *RealioNetwork) ScheduleForkUpgrade(ctx sdk.Context) {
	if ctx.BlockHeight() == ForkHeight {

		// remove duplicate UnbondingQueueKey
		removeDuplicateValueUnbondingQueueKey(app, ctx)
		removeDuplicateValueRedelegationQueueKey(app, ctx)
		removeDuplicateUnbondingValidator(app, ctx)
	}
	// NOTE: there are no testnet forks for the existing versions
	// if !types.IsMainnet(ctx.ChainID()) {
	//	return
	//}
	//
	// upgradePlan := upgradetypes.Plan{
	//	Height: ctx.BlockHeight(),
	//}
	//
	//// handle mainnet forks with their corresponding upgrade name and info
	// switch ctx.BlockHeight() {
	// case v2.MainnetUpgradeHeight:
	//	upgradePlan.Name = v2.UpgradeName
	//	upgradePlan.Info = v2.UpgradeInfo
	//default:
	//	// No-op
	//	return
	//}
	//
	//// schedule the upgrade plan to the current block height, effectively performing
	//// a hard fork that uses the upgrade handler to manage the migration.
	// if err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan); err != nil {
	//	panic(
	//		fmt.Errorf(
	//			"failed to schedule upgrade %s during BeginBlock at height %d: %w",
	//			upgradePlan.Name, ctx.BlockHeight(), err,
	//		),
	//	)
	//}
}

func removeDuplicateValueRedelegationQueueKey(app *RealioNetwork, ctx sdk.Context) {
	// Get Staking keeper, codec and staking store
	sk := app.StakingKeeper
	cdc := app.AppCodec()
	store := ctx.KVStore(app.keys[stakingtypes.ModuleName])

	redelegationTimesliceIterator := sk.RedelegationQueueIterator(ctx, oneEnternityLater) // make sure to iterate all queue
	defer redelegationTimesliceIterator.Close()

	for ; redelegationTimesliceIterator.Valid(); redelegationTimesliceIterator.Next() {
		timeslice := stakingtypes.DVVTriplets{}
		value := redelegationTimesliceIterator.Value()
		cdc.MustUnmarshal(value, &timeslice)

		triplets := removeDuplicateDVVTriplets(timeslice.Triplets)
		bz := cdc.MustMarshal(&stakingtypes.DVVTriplets{Triplets: triplets})

		store.Set(redelegationTimesliceIterator.Key(), bz)
	}
}

func removeDuplicateDVVTriplets(triplets []stakingtypes.DVVTriplet) []stakingtypes.DVVTriplet {
	var list []stakingtypes.DVVTriplet
	for _, item := range triplets {
		if !containsDVVTriplets(list, item) {
			list = append(list, item)
		}
	}
	return list
}

func containsDVVTriplets(s []stakingtypes.DVVTriplet, e stakingtypes.DVVTriplet) bool {
	for _, a := range s {
		if a.DelegatorAddress == e.DelegatorAddress &&
			a.ValidatorSrcAddress == e.ValidatorSrcAddress &&
			a.ValidatorDstAddress == e.ValidatorDstAddress {
			return true
		}
	}
	return false
}

func removeDuplicateUnbondingValidator(app *RealioNetwork, ctx sdk.Context) {
	valIter := app.StakingKeeper.ValidatorQueueIterator(ctx, oneEnternityLater, 99999999999999)
	defer valIter.Close()

	for ; valIter.Valid(); valIter.Next() {
		addrs := stakingtypes.ValAddresses{}
		app.appCodec.MustUnmarshal(valIter.Value(), &addrs)

		vals := map[string]bool{}
		for _, valAddr := range addrs.Addresses {
			vals[valAddr] = true
		}

		uniqueAddrs := []string{}
		for valAddr := range vals {
			uniqueAddrs = append(uniqueAddrs, valAddr)
		}
		sort.Strings(uniqueAddrs)

		ctx.KVStore(app.GetKey(stakingtypes.StoreKey)).Set(valIter.Key(), app.appCodec.MustMarshal(&stakingtypes.ValAddresses{Addresses: uniqueAddrs}))
	}
}

func removeDuplicateValueUnbondingQueueKey(app *RealioNetwork, ctx sdk.Context) {
	// Get Staking keeper, codec and staking store
	sk := app.StakingKeeper
	cdc := app.AppCodec()
	store := ctx.KVStore(app.keys[stakingtypes.ModuleName])

	unbondingTimesliceIterator := sk.UBDQueueIterator(ctx, oneEnternityLater) // make sure to iterate all queue
	defer unbondingTimesliceIterator.Close()

	for ; unbondingTimesliceIterator.Valid(); unbondingTimesliceIterator.Next() {
		timeslice := stakingtypes.DVPairs{}
		value := unbondingTimesliceIterator.Value()
		cdc.MustUnmarshal(value, &timeslice)

		dvPairs := removeDuplicatesDVPairs(timeslice.Pairs)
		bz := cdc.MustMarshal(&stakingtypes.DVPairs{Pairs: dvPairs})

		store.Set(unbondingTimesliceIterator.Key(), bz)
	}
}

func removeDuplicatesDVPairs(dvPairs []stakingtypes.DVPair) []stakingtypes.DVPair {
	var list []stakingtypes.DVPair
	for _, item := range dvPairs {
		if !containsDVPairs(list, item) {
			list = append(list, item)
		}
	}
	return list
}

func containsDVPairs(s []stakingtypes.DVPair, e stakingtypes.DVPair) bool {
	for _, a := range s {
		if a.DelegatorAddress == e.DelegatorAddress &&
			a.ValidatorAddress == e.ValidatorAddress {
			return true
		}
	}
	return false
}

func (app *RealioNetwork) ScheduleForkUpgrade1(ctx sdk.Context) {
	upgradePlan := upgradetypes.Plan{
		Height: ctx.BlockHeight()+10,
	}

	// handle mainnet forks with their corresponding upgrade name and info
	switch ctx.BlockHeight()+10 {
	case 7302000:
		upgradePlan.Name = "v2"
	default:
		// No-op
		return
	}

	// schedule the upgrade plan to the current block height, effectively performing
	// a hard fork that uses the upgrade handler to manage the migration.
	if err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan); err != nil {
		panic(
			fmt.Errorf(
				"failed to schedule upgrade %s during BeginBlock at height %d: %w",
				upgradePlan.Name, ctx.BlockHeight(), err,
			),
		)
	}
}