package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	core "github.com/terra-project/core/types"
	"github.com/terra-project/core/x/market/internal/types"
)

// Simulation parameter constants
const (
	basePoolKey             = "base_pool"
	poolRecoveryPeriodKey   = "pool_recovery_period"
	minSpreadKey            = "min_spread"
	tobinTaxKey             = "tobin_tax"
	illiquidTobinTaxRateKey = "illiquid_tobin_tax_rate"
)

// GenBasePool randomized BasePool
func GenBasePool(r *rand.Rand) sdk.Dec {
	return sdk.NewDec(100000000).Add(sdk.NewDec(int64(r.Intn(10000000000))))
}

// GenPoolRecoveryPeriod randomized PoolRecoveryPeriod
func GenPoolRecoveryPeriod(r *rand.Rand) int64 {
	return int64(100 + r.Intn(10000000000))
}

// GenMinSpread randomized MinSpread
func GenMinSpread(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(100)), 3))
}

// GenTobinTax randomized TobinTax
func GenTobinTax(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1, 3).Add(sdk.NewDecWithPrec(int64(r.Intn(100)), 3))
}

// GenIlliquidTobinTaxRate randomized IlliquidTobinTaxRate
func GenIlliquidTobinTaxRate(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(100)), 3))
}

// RandomizedGenState generates a random GenesisState for gov
func RandomizedGenState(simState *module.SimulationState) {

	var basePool sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, basePoolKey, &basePool, simState.Rand,
		func(r *rand.Rand) { basePool = GenBasePool(r) },
	)

	var poolRecoveryPeriod int64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, poolRecoveryPeriodKey, &poolRecoveryPeriod, simState.Rand,
		func(r *rand.Rand) { poolRecoveryPeriod = GenPoolRecoveryPeriod(r) },
	)

	var minSpread sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, minSpreadKey, &minSpread, simState.Rand,
		func(r *rand.Rand) { minSpread = GenMinSpread(r) },
	)

	var tobinTax sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, tobinTaxKey, &tobinTax, simState.Rand,
		func(r *rand.Rand) { tobinTax = GenTobinTax(r) },
	)

	var illiquidTobinTaxRate sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, illiquidTobinTaxRateKey, &illiquidTobinTaxRate, simState.Rand,
		func(r *rand.Rand) { illiquidTobinTaxRate = GenIlliquidTobinTaxRate(r) },
	)

	marketGenesis := types.NewGenesisState(
		sdk.ZeroDec(),
		types.Params{
			BasePool:           basePool,
			PoolRecoveryPeriod: poolRecoveryPeriod,
			MinSpread:          minSpread,
			TobinTax:           tobinTax,
			IlliquidTobinTaxList: types.TobinTaxList{
				{Denom: core.MicroMNTDenom, TaxRate: illiquidTobinTaxRate},
			},
		},
	)

	fmt.Printf("Selected randomly generated market parameters:\n%s\n", codec.MustMarshalJSONIndent(simState.Cdc, marketGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(marketGenesis)
}
