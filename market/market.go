package market

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	types "github.com/cosmos/cosmos-sdk/types"
)

func NewInterchainLiquidityPool(
	creator string,
	tokens []*types.DecCoin,
	decimals []uint32,
	weight string,
	portId string,
	channelId string,
) *InterchainLiquidityPool {

	//generate poolId
	poolId := GetPoolIdWithTokens(tokens)

	weights := strings.Split(weight, ":")
	weightSize := len(weights)
	denomSize := len(tokens)
	decimalSize := len(decimals)
	assets := []*PoolAsset{}

	if denomSize == weightSize && decimalSize == weightSize {
		for index, token := range tokens {
			side := PoolSide_NATIVE
			// if !store.HasSupply(ctx, token.Denom) {
			// 	side = PoolSide_REMOTE
			// }
			weight, _ := strconv.ParseUint(weights[index], 10, 32)
			asset := PoolAsset{
				Side:    side,
				Balance: token,
				Weight:  uint32(weight),
				Decimal: decimals[index],
			}
			assets = append(assets, &asset)
		}
	} else {
		return nil
	}
	// amm := NewInterchainMarketMaker(

	// )
	return &InterchainLiquidityPool{
		PoolId:  poolId,
		Creator: creator,
		Assets:  assets,
		Supply: &types.DecCoin{
			Amount: types.NewDec(0),
			Denom:  poolId,
		},
		Status:                PoolStatus_POOL_STATUS_READY,
		EncounterPartyPort:    portId,
		EncounterPartyChannel: channelId,
	}
}

// find pool asset by denom
func (ilp *InterchainLiquidityPool) FindAssetByDenom(denom string) (*PoolAsset, error) {
	for _, asset := range ilp.Assets {
		if asset.Balance.Denom == denom {
			return asset, nil
		}
	}
	return nil, fmt.Errorf("did not find asset")
}

// update denom
func (ilp *InterchainLiquidityPool) UpdateAssetPoolSide(denom string, side PoolSide) (*PoolAsset, error) {
	for index, asset := range ilp.Assets {
		if asset.Balance.Denom == denom {
			ilp.Assets[index].Side = side
		}
	}
	return nil, fmt.Errorf("did not find asset")
}

// update denom
func (ilp *InterchainLiquidityPool) AddAsset(token types.DecCoin) error {
	for index, asset := range ilp.Assets {
		if asset.Balance.Denom == token.Denom {
			updatedCoin := ilp.Assets[index].Balance.Add(token)
			ilp.Assets[index].Balance = &updatedCoin
		}
	}
	return fmt.Errorf("did not find asset")
}

// update denom
func (ilp *InterchainLiquidityPool) SubAsset(token types.DecCoin) error {
	for index, asset := range ilp.Assets {
		if asset.Balance.Denom == token.Denom {
			updatedCoin := ilp.Assets[index].Balance.Sub(token)
			ilp.Assets[index].Balance = &updatedCoin
		}
	}
	return fmt.Errorf("did not find asset")
}

// update pool suppy
func (ilp *InterchainLiquidityPool) AddPoolSupply(token types.DecCoin) error {
	if token.Denom != ilp.PoolId {
		return fmt.Errorf("invalid denom")
	}
	updatedCoin := ilp.Supply.Add(token)
	ilp.Supply = &updatedCoin
	return nil
}

// update pool suppy
func (ilp *InterchainLiquidityPool) SubPoolSupply(token types.DecCoin) error {
	if token.Denom != ilp.PoolId {
		return fmt.Errorf("invalid denom")
	}
	updatedCoin := ilp.Supply.Sub(token)
	ilp.Supply = &updatedCoin
	return nil
}

//create new market maker

func NewInterchainMarketMaker(
	pool *InterchainLiquidityPool,
	feeRate uint32,
) *InterchainMarketMaker {
	return &InterchainMarketMaker{
		Pool:    pool,
		FeeRate: feeRate,
	}
}

// MarketPrice Bi / Wi / (Bo / Wo)
func (imm *InterchainMarketMaker) MarketPrice(denomIn, denomOut string) (*types.Dec, error) {
	tokenIn, err := imm.Pool.FindAssetByDenom(denomIn)
	if err != nil {
		return nil, err
	}

	tokenOut, err := imm.Pool.FindAssetByDenom(denomOut)
	if err != nil {
		return nil, err
	}

	balanceIn := tokenIn.Balance.Amount
	balanceOut := tokenOut.Balance.Amount
	weightIn := tokenIn.Weight
	weightOut := tokenOut.Weight

	// Convert all values to Dec type
	balanceInDec := types.NewDecFromBigInt(balanceIn.BigInt())
	balanceOutDec := types.NewDecFromBigInt(balanceOut.BigInt())
	weightInDec := types.NewDecFromInt(types.NewInt(int64(weightIn)))
	weightOutDec := types.NewDecFromInt(types.NewInt(int64(weightOut)))

	// Perform calculations using Dec type
	ratioIn := balanceInDec.Quo(weightInDec)
	ratioOut := balanceOutDec.Quo(weightOutDec)
	marketPrice := ratioIn.Quo(ratioOut)

	return &marketPrice, nil
}

// Worth Function V=M)
func (imm *InterchainMarketMaker) Worth() (*types.Dec, error) {
	v := 1.0
	for _, pool := range imm.Pool.Assets {
		w := float64(pool.Weight) / 100.0
		balance := pool.Balance.Amount.MustFloat64()
		v *= math.Pow(balance, w)
	}
	decV := types.NewDec(int64(v))
	return &decV, nil
}

// Worth Function V=M)
func (imm *InterchainMarketMaker) TVL() float64 {
	v := 1.0
	totalBalance := types.NewDec(0)
	for _, pool := range imm.Pool.Assets {
		totalBalance = totalBalance.Add(pool.Balance.Amount)
		//totalBalance.Add(totalBalance, pool.Balance.Amount.BigInt())
	}
	for _, pool := range imm.Pool.Assets {
		w := float64(pool.Weight) / 100.0
		balance := pool.Balance.Amount.MustFloat64() / totalBalance.MustFloat64()
		v *= math.Pow(balance, w)
	}

	return v
}

func (imm *InterchainMarketMaker) DepositSingleAsset(token types.DecCoin) (*types.DecCoin, error) {
	asset, err := imm.Pool.FindAssetByDenom(token.Denom)
	if err != nil {
		return nil, err
	}

	// // Get the maximum single deposit amount based on the current weight imbalance
	// weightImbalance := imm.getWeightImbalance()
	// maxDepositAmount := types.NewDecFromInt(asset.Balance.Amount).Mul(weightImbalance).RoundInt()

	// if token.Amount.GT(maxDepositAmount) {
	// 	return nil, fmt.Errorf("single deposit amount is greater than the allowed limit based on the current weight imbalance")
	// }

	//amountDec := token.Amount           //types.NewDecFromBigInt(token.Amount.BigInt())
	//supplyDec := imm.Pool.Supply.Amount //types.NewDecFromBigInt(imm.Pool.Supply.Amount.BigInt())
	// weightDec := types.NewDecFromInt(types.NewInt(int64(asset.Weight))).Quo(types.NewDec(100))

	var issueAmount types.Dec
	if imm.Pool.Status == PoolStatus_POOL_STATUS_INITIAL {
		totalInitialLpAmount := types.NewDec(0)
		for _, asset := range imm.Pool.Assets {
			totalInitialLpAmount = totalInitialLpAmount.Add(asset.Balance.Amount)
		}
		issueAmount = totalInitialLpAmount.Mul(types.NewDec(int64(asset.Weight))).Quo(types.NewDec(100))
	} else {
		weight := float64(asset.Weight) / 100
		ratio := 1 + float64(token.Amount.MustFloat64())/float64(asset.Balance.Amount.MustFloat64())
		factor := math.Pow(ratio, float64(weight)) - 1
		issueLp := float64(imm.Pool.Supply.Amount.MustFloat64()) * factor
		issueAmount = types.NewDec(int64(issueLp))
		// balanceDec := asset.Balance.Amount
		// ratio := amountDec.Quo(balanceDec)

		// // Calculate log(1 + ratio)
		// logFactorBase := types.NewDec(1).Add(ratio).MustFloat64()
		// logValue := math.Log(logFactorBase)

		// // Calculate logValue * weight
		// weightFloat64 := weightDec.MustFloat64()
		// logTimesWeight := logValue * weightFloat64

		// // Calculate e^(logTimesWeight) - 1
		// factor := math.Exp(logTimesWeight) - 1

		// // Convert float64 back to Dec
		// factorDec := types.NewDecFromInt(types.NewInt(int64(factor * 1e18))).Quo(types.NewDecFromInt(types.NewInt(1e18)))

		// issueAmount = supplyDec.Mul(factorDec)
		// //issueAmount = types.NewIntFromBigInt(issueAmountDec.TruncateInt().BigInt())
	}

	outputToken := &types.DecCoin{
		Amount: issueAmount,
		Denom:  imm.Pool.Supply.Denom,
	}

	// pool status update.
	imm.Pool.AddAsset(token)
	imm.Pool.AddPoolSupply(*outputToken)

	// adjust pool weight
	//imm.updateWeights()

	return outputToken, nil
}

func (imm *InterchainMarketMaker) Withdraw(redeem types.DecCoin, denomOut string) (*types.Coin, error) {
	asset, err := imm.Pool.FindAssetByDenom(denomOut)
	if err != nil {
		return nil, err
	}
	err = asset.Balance.Validate()
	if err != nil {
		return nil, err
	}
	if imm.Pool.Status != PoolStatus_POOL_STATUS_READY {
		return nil, fmt.Errorf("not ready for swap")
	}

	if redeem.Amount.GT(imm.Pool.Supply.Amount) {
		return nil, fmt.Errorf("bigger than balance")
	}

	if redeem.Denom != imm.Pool.Supply.Denom {
		return nil, fmt.Errorf("invalid denom pair")
	}

	// At = Bt * (1 - (1 - P_redeemed / P_supply) ** 1/Wt)

	//asset.Balance.Amount
	// ratio := 1 - float64(redeem.Amount.Uint64())/float64(imm.Pool.Supply.Amount.Uint64())
	// exponent := 1 / float64(asset.Weight)
	// factor := 1 - math.Pow(ratio, exponent)

	balance := asset.Balance.Amount
	supply := imm.Pool.Supply.Amount
	weight := types.NewDec(int64(asset.Weight)).Quo(types.NewDec(100))

	// // At = Bt * (1 - (1 - P_redeemed / P_supply) ** 1/Wt)
	redeemDec := redeem.Amount
	oneMinusRatio := types.NewDec(1).Sub(redeemDec.Quo(supply))

	// Calculate log(1 - ratio)
	logOneMinusRatio, _ := oneMinusRatio.Float64()
	logValue := math.Log(logOneMinusRatio)

	// Calculate logValue / weight
	weightFloat64, _ := weight.Float64()
	logDividedByWeight := logValue / weightFloat64

	// Calculate 1 - e^(logDividedByWeight)
	factor := 1 - math.Exp(logDividedByWeight)

	// Convert float64 back to Dec
	factorDec := types.NewDecFromInt(types.NewInt(int64(factor * 1e18))).Quo(types.NewDecFromInt(types.NewInt(1e18)))
	amountOut := balance.Mul(factorDec)
	return &types.Coin{
		Amount: amountOut.RoundInt(),
		Denom:  denomOut,
	}, nil
}

// LeftSwap implements OutGivenIn
// Input how many coins you want to sell, output an amount you will receive
// Ao = Bo * ((1 - Bi / (Bi + Ai)) ** Wi/Wo)
func (imm *InterchainMarketMaker) LeftSwap(amountIn types.DecCoin, denomOut string) (*types.DecCoin, error) {
	assetIn, err := imm.Pool.FindAssetByDenom(amountIn.Denom)
	if err != nil {
		return nil, err
	}

	assetOut, err := imm.Pool.FindAssetByDenom(denomOut)
	if err != nil {
		return nil, err
	}

	balanceOut := assetOut.Balance.Amount
	balanceIn := assetIn.Balance.Amount
	weightIn := types.NewDec(int64(assetIn.Weight)).Quo(types.NewDec(100))
	weightOut := types.NewDec(int64(assetOut.Weight)).Quo(types.NewDec(100))
	amount := imm.MinusFees(amountIn.Amount)

	// Ao = Bo * ((1 - Bi / (Bi + Ai)) ** Wi/Wo)
	balanceInPlusAmount := balanceIn.Add(amount)
	ratio := balanceIn.Quo(balanceInPlusAmount)
	oneMinusRatio := types.NewDec(1).Sub(ratio)
	power := weightIn.Quo(weightOut)
	factor := math.Pow(oneMinusRatio.MustFloat64(), power.MustFloat64()) * 1e18
	amountOut := balanceOut.Mul(types.NewDec(int64(factor))).Quo(types.NewDec(1e18))

	return &types.DecCoin{
		Amount: amountOut,
		Denom:  denomOut,
	}, nil
}

// RightSwap implements InGivenOut
// Input how many coins you want to buy, output an amount you need to pay
// Ai = Bi * ((Bo/(Bo - Ao)) ** Wo/Wi -1)
func (imm *InterchainMarketMaker) RightSwap(amountIn types.DecCoin, amountOut types.Coin) (*types.DecCoin, error) {
	assetIn, err := imm.Pool.FindAssetByDenom(amountIn.Denom)
	if err != nil {
		return nil, fmt.Errorf("right swap failed")
	}

	assetOut, err := imm.Pool.FindAssetByDenom(amountOut.Denom)
	if err != nil {
		return nil, fmt.Errorf("right swap failed")
	}

	//Ai = Bi * ((Bo/(Bo - Ao)) ** Wo/Wi -1)
	balanceIn := assetIn.Balance.Amount
	weightIn := types.NewDec(int64(assetIn.Weight)).Quo(types.NewDec(100))
	weightOut := types.NewDec(int64(assetOut.Weight)).Quo(types.NewDec(100))

	//one := types.NewDec(1)
	numerator := assetOut.Balance.Amount
	power := weightOut.Quo(weightIn)
	denominator := assetOut.Balance.Amount.Sub(assetIn.Balance.Amount)
	base := numerator.Quo(denominator)
	factor := math.Pow(base.MustFloat64(), power.MustFloat64()) * 1e18
	amountRequired := balanceIn.Mul(types.NewDec(int64(factor))).Quo(types.NewDec(1e18))

	if amountIn.Amount.LT(amountRequired) {
		return nil, fmt.Errorf("right swap failed")
	}
	return &types.DecCoin{
		Amount: amountRequired,
		Denom:  amountIn.Denom,
	}, nil
}

func (imm *InterchainMarketMaker) MinusFees(amount types.Dec) types.Dec {
	feeRate := types.NewDec(int64(imm.FeeRate)).Quo(types.NewDec(10000))
	fees := amount.Mul(feeRate)
	amountMinusFees := amount.Sub(fees)
	return amountMinusFees
}

func (amm *InterchainMarketMaker) LogPrice(
	title string,
	denomA string,
	denomB string,
	userALp, userBLp types.Int,
) {
	fmt.Println("---------------", title, "----------------")
	fmt.Println("lp supply amount:", amm.Pool.Supply.Amount)
	for _, asset := range amm.Pool.Assets {
		fmt.Println(asset.Balance.Denom, ":", asset.Balance.Amount)
	}

	for _, asset := range amm.Pool.Assets {
		fmt.Println(asset.Balance.Denom, ":", asset.Balance.Amount)
	}
	worth := amm.TVL()
	lpPrice := worth * 1e10 / amm.Pool.Supply.Amount.MustFloat64()

	outA, _ := amm.Withdraw(types.NewDecCoin(amm.Pool.PoolId, userALp), denomA)
	outB, _ := amm.Withdraw(types.NewDecCoin(amm.Pool.PoolId, userBLp), denomB)

	fmt.Println("----*********----")
	fmt.Println("UserALPAmount:", userALp)
	fmt.Println("UserBLPAmount:", userBLp)
	fmt.Println("WithdrawUSDT:", outA)
	fmt.Println("WithdrawETH:", outB)
	fmt.Println("Worth:", worth)
	fmt.Println("LPprice:", lpPrice)
	// fmt.Println("--------------------------------------------\n")
}

func (imm *InterchainMarketMaker) CheckMaxDepositAmount(tokenDenom string, maxLPPriceChangePercent float64) (*types.DecCoin, error) {
	asset, err := imm.Pool.FindAssetByDenom(tokenDenom)
	if err != nil {
		return nil, err
	}

	// Calculate the initial LP price
	initialLpPrice := imm.Pool.PoolTokenPrice

	// Calculate the initial LP price
	currentWorth := imm.TVL()
	currentLpPrice := currentWorth * 1e10 / imm.Pool.Supply.Amount.MustFloat64()

	maxLpPrice := float64(0)
	priceImbalance := math.Abs(initialLpPrice - currentLpPrice)
	if priceImbalance > 0 {
		maxLpPrice = initialLpPrice * (1 + (currentLpPrice-initialLpPrice)/100)
	} else {
		maxLpPrice = initialLpPrice * (1 + maxLPPriceChangePercent/100)
	}

	// Find the maximum deposit amount using the formula derived above
	weight := float64(asset.Weight)
	balance := float64(asset.Balance.Amount.MustFloat64())
	maxDeposit := math.Pow(maxLpPrice/initialLpPrice, 100/weight)*balance - balance
	return &types.DecCoin{
		Amount: types.NewDec(int64(maxDeposit)),
		Denom:  tokenDenom,
	}, nil
}
