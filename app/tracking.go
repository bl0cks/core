package app

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	core "github.com/terra-project/core/types"
	"github.com/terra-project/core/x/auth"
	"github.com/terra-project/core/x/staking"
)

func (app *TerraApp) trackingAll(ctx sdk.Context) {
	// Build validator token share map to calculate delegators staking tokens
	validators := staking.Validators(app.stakingKeeper.GetAllValidators(ctx))
	tokenShareRates := make(map[string]sdk.Dec)
	for _, validator := range validators {
		tokenShareRates[validator.GetOperator().String()] = validator.GetBondedTokens().ToDec().Quo(validator.GetDelegatorShares())
	}

	// Load oracle whitelist
	denoms := app.oracleKeeper.Whitelist(ctx)

	// Minimum coins to be included in tracking
	minCoins := sdk.Coins{}
	for _, denom := range denoms {
		minCoins = append(minCoins, sdk.NewCoin(denom, sdk.OneInt().MulRaw(core.MicroUnit)))
	}

	minCoins = minCoins.Sort()

	accs := []authexported.Account{}
	vestingCoins := sdk.NewCoins()
	app.accountKeeper.IterateAccounts(ctx, func(acc authexported.Account) bool {
		// Record vesting accounts
		if vacc, ok := acc.(auth.VestingAccount); ok {
			vestingCoins = vestingCoins.Add(vacc.GetVestingCoins(ctx.BlockHeader().Time))
		}

		// Compute staking amount
		stakingAmt := sdk.ZeroInt()
		delegations := app.stakingKeeper.GetAllDelegatorDelegations(ctx, acc.GetAddress())
		undelegations := app.stakingKeeper.GetUnbondingDelegations(ctx, acc.GetAddress(), 100)
		for _, delegation := range delegations {
			valAddr := delegation.GetValidatorAddr().String()
			tokenShareRate := tokenShareRates[valAddr]
			delegationAmt := delegation.GetShares().Mul(tokenShareRate).TruncateInt()

			stakingAmt = stakingAmt.Add(delegationAmt)
		}

		unbondingAmt := sdk.ZeroInt()
		for _, undelegation := range undelegations {
			undelegationAmt := sdk.ZeroInt()
			for _, entry := range undelegation.Entries {
				undelegationAmt = undelegationAmt.Add(entry.Balance)
			}

			unbondingAmt.Add(undelegationAmt)
		}

		// Add staking amount to account balance
		stakingCoins := sdk.NewCoins(sdk.NewCoin(app.stakingKeeper.BondDenom(ctx), stakingAmt.Add(unbondingAmt)))
		acc.SetCoins(acc.GetCoins().Add(stakingCoins))

		// Check minimum coins
		if acc.GetCoins().IsAnyGTE(minCoins) {
			accs = append(accs, acc)
		}

		return false
	})

	go app.exportVestingSupply(ctx, vestingCoins)
	go app.exportRanking(ctx, accs, denoms)

}

func (app *TerraApp) exportVestingSupply(ctx sdk.Context, vestingCoins sdk.Coins) {
	app.Logger().Info("Start Tracking Vesting Luna Supply")
	bz, err := codec.MarshalJSONIndent(app.cdc, vestingCoins)
	if err != nil {
		app.Logger().Error(err.Error())
	}

	err = ioutil.WriteFile(fmt.Sprintf("/tmp/vesting-%s.json", time.Now().Format(time.RFC3339)), bz, 0644)
	if err != nil {
		app.Logger().Error(err.Error())
	}
	app.Logger().Info("End Tracking Vesting Luna Supply")
}

// ExportAccount is ranking export account format
type ExportAccount struct {
	Address sdk.AccAddress `json:"address"`
	Amount  sdk.Int        `json:"amount"`
}

// NewExportAccount returns new ExportAccount instance
func NewExportAccount(address sdk.AccAddress, amount sdk.Int) ExportAccount {
	return ExportAccount{
		Address: address,
		Amount:  amount,
	}
}

func (app *TerraApp) exportRanking(ctx sdk.Context, accs []auth.Account, denoms []string) {
	app.Logger().Info("Start Tracking Top 1000 Rankers")

	maxEntries := 1000
	if len(accs) < maxEntries {
		maxEntries = len(accs)
	}

	for _, denom := range denoms {

		var topRankerList []ExportAccount

		tmpAccs := make([]auth.Account, len(accs))
		copy(tmpAccs, accs)

		for i := 0; i < maxEntries; i++ {

			var topRankerAmt sdk.Int
			var topRankerAddr sdk.AccAddress
			var topRankerIdx int

			for idx, acc := range tmpAccs {
				addr := acc.GetAddress()
				amt := acc.GetCoins().AmountOf(denom)

				if idx == 0 || amt.GT(topRankerAmt) {
					topRankerIdx = idx
					topRankerAmt = amt
					topRankerAddr = addr
				}
			}

			topRankerList = append(topRankerList, NewExportAccount(topRankerAddr, topRankerAmt))
			tmpAccs[topRankerIdx] = tmpAccs[len(tmpAccs)-1]
			tmpAccs = tmpAccs[:len(tmpAccs)-1]
		}

		bz, err := codec.MarshalJSONIndent(app.cdc, topRankerList)
		if err != nil {
			app.Logger().Error(err.Error())
		}

		err = ioutil.WriteFile(fmt.Sprintf("/tmp/tracking-%s-%s.json", denom, time.Now().Format(time.RFC3339)), bz, 0644)
		if err != nil {
			app.Logger().Error(err.Error())
		}
	}

	app.Logger().Info("End Tracking Top 1000 Rankers")
}