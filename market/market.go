package market

import (
	"fmt"
	math "math"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
)

const Multiplier = 1e18

// create new liquidity pool
func NewInterchainLiquidityPool(
	creator string,
	tokens []*types.Coin,
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
		Supply: &types.Coin{
			Amount: types.NewInt(0),
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
	return nil, fmt.Errorf("did not find denom")
}

// update denom
func (ilp *InterchainLiquidityPool) UpdateAssetPoolSide(denom string, side PoolSide) (*PoolAsset, error) {
	for index, asset := range ilp.Assets {
		if asset.Balance.Denom == denom {
			ilp.Assets[index].Side = side
		}
	}
	return nil, fmt.Errorf("did not find denom")
}

// update denom
func (ilp *InterchainLiquidityPool) AddAsset(token types.Coin) error {
	for index, asset := range ilp.Assets {
		if asset.Balance.Denom == token.Denom {
			updatedCoin := ilp.Assets[index].Balance.Add(token)
			ilp.Assets[index].Balance = &updatedCoin
		}
	}
	return fmt.Errorf("did not find denom")
}

// update denom
func (ilp *InterchainLiquidityPool) SubAsset(token types.Coin) error {
	for index, asset := range ilp.Assets {
		if asset.Balance.Denom == token.Denom {
			updatedCoin := ilp.Assets[index].Balance.Sub(token)
			ilp.Assets[index].Balance = &updatedCoin
		}
	}
	return fmt.Errorf("did not find denom")
}

// update pool suppy
func (ilp *InterchainLiquidityPool) AddPoolSupply(token types.Coin) error {
	if token.Denom != ilp.PoolId {
		return fmt.Errorf("did not find denom")
	}
	updatedCoin := ilp.Supply.Add(token)
	ilp.Supply = &updatedCoin
	return nil
}

// update pool suppy
func (ilp *InterchainLiquidityPool) SubPoolSupply(token types.Coin) error {
	if token.Denom != ilp.PoolId {
		return fmt.Errorf("did not find denom")
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

// P_issued = P_supply * ((1 + At/Bt) ** Wt -1)
func (imm *InterchainMarketMaker) DepositSingleAsset(token types.Coin) (*types.Coin, error) {

	asset, err := imm.Pool.FindAssetByDenom(token.Denom)
	if err != nil {
		return nil, err
	}

	var issueAmount types.Int

	if imm.Pool.Status == PoolStatus_POOL_STATUS_INITIAL {
		totalInitialLpAmount := types.NewInt(0)
		for _, asset := range imm.Pool.Assets {
			totalInitialLpAmount = totalInitialLpAmount.Add(asset.Balance.Amount)
		}
		issueAmount = totalInitialLpAmount
	} else {

		weight := float64(asset.Weight) / 100
		ratio := 1 + float64(token.Amount.Quo(asset.Balance.Amount).Int64())
		factor := (math.Pow(ratio, float64(weight)) - 1) * Multiplier
		issueAmount = imm.Pool.Supply.Amount.Mul(types.NewInt(int64(factor))).Quo(types.NewInt(Multiplier))

		estimatedAmount := imm.Pool.Supply.Amount.Add(issueAmount)
		estimatedLpPrice := imm.InvariantWithInput(token) / float64(estimatedAmount.Int64())
		if math.Abs(estimatedLpPrice-float64(imm.Pool.PoolPrice))/float64(imm.Pool.PoolPrice) > 0.1 {
			return nil, fmt.Errorf("this deposit break pool equbrillam")
		}
	}

	outputToken := &types.Coin{
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

// input the supply token, output the expected token.
// At = Bt * (1 - (1 - P_redeemed / P_supply) ** 1/Wt)
func (imm *InterchainMarketMaker) Withdraw(redeem types.Coin, denomOut string) (*types.Coin, error) {
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

	ratio := 1 - float64(redeem.Amount.Mul(types.NewInt(Multiplier)).Quo(imm.Pool.Supply.Amount).Int64())/Multiplier
	exponent := 1 / float64(asset.Weight)
	factor := (1 - math.Pow(ratio, exponent)) * Multiplier
	amountOut := asset.Balance.Amount.Mul(types.NewInt(int64(factor))).Quo(types.NewInt(Multiplier))
	return &types.Coin{
		Amount: amountOut,
		Denom:  denomOut,
	}, nil
}

// LeftSwap implements OutGivenIn
// Input how many coins you want to sell, output an amount you will receive
// Ao = Bo * ((1 - Bi / (Bi + Ai)) ** Wi/Wo)
func (imm *InterchainMarketMaker) LeftSwap(amountIn types.Coin, denomOut string) (*types.Coin, error) {
	assetIn, err := imm.Pool.FindAssetByDenom(amountIn.Denom)
	if err != nil {
		return nil, err
	}

	assetOut, err := imm.Pool.FindAssetByDenom(denomOut)
	if err != nil {
		return nil, err
	}

	balanceOut := types.NewDecFromBigInt(assetOut.Balance.Amount.BigInt())
	balanceIn := types.NewDecFromBigInt(assetIn.Balance.Amount.BigInt())
	weightIn := types.NewDec(int64(assetIn.Weight)).Quo(types.NewDec(100))
	weightOut := types.NewDec(int64(assetOut.Weight)).Quo(types.NewDec(100))
	amount := imm.MinusFees(amountIn.Amount)

	// Ao = Bo * (1 - (Bi / (Bi + Ai)) ** Wi/Wo)
	balanceInPlusAmount := balanceIn.Add(amount)
	ratio := balanceIn.Quo(balanceInPlusAmount)
	power := weightIn.Quo(weightOut)
	math.Pow(ratio.MustFloat64(), power.MustFloat64())
	factor := (1 - math.Pow(ratio.MustFloat64(), power.MustFloat64())) * 1e18
	amountOut := balanceOut.Mul(types.NewDec(int64(factor))).Quo(types.NewDec(1e18))

	fmt.Println(
		"left swap params:", power, factor, amountOut,
	)
	return &types.Coin{
		Amount: amountOut.RoundInt(),
		Denom:  denomOut,
	}, nil
}

// RightSwap implements InGivenOut
// Input how many coins you want to buy, output an amount you need to pay
// Ai = Bi * ((Bo/(Bo - Ao)) ** Wo/Wi -1)
func (imm *InterchainMarketMaker) RightSwap(amountIn types.Coin, amountOut types.Coin) (*types.Coin, error) {
	assetIn, err := imm.Pool.FindAssetByDenom(amountIn.Denom)
	if err != nil {
		return nil, fmt.Errorf("right swap failed")
	}

	assetOut, err := imm.Pool.FindAssetByDenom(amountOut.Denom)
	if err != nil {
		return nil, fmt.Errorf("right swap failed")
	}

	//Ai = Bi * ((Bo/(Bo - Ao)) ** Wo/Wi -1)
	balanceIn := types.NewDecFromBigInt(assetIn.Balance.Amount.BigInt())
	weightIn := types.NewDec(int64(assetIn.Weight)).Quo(types.NewDec(100))
	weightOut := types.NewDec(int64(assetOut.Weight)).Quo(types.NewDec(100))

	//one := types.NewDec(1)
	numerator := types.NewDecFromBigInt(assetOut.Balance.Amount.BigInt())
	power := weightOut.Quo(weightIn)
	denominator := types.NewDecFromBigInt(assetOut.Balance.Amount.Sub(assetIn.Balance.Amount).BigInt())
	base := numerator.Quo(denominator)
	factor := math.Pow(base.MustFloat64(), power.MustFloat64()) * Multiplier
	amountRequired := balanceIn.Mul(types.NewDec(int64(factor))).Quo(types.NewDec(Multiplier)).RoundInt()

	if amountIn.Amount.LT(amountRequired) {
		return nil, fmt.Errorf("right swap failed")
	}
	return &types.Coin{
		Amount: amountRequired,
		Denom:  amountIn.Denom,
	}, nil
}

func (imm *InterchainMarketMaker) MinusFees(amount types.Int) types.Dec {
	amountDec := types.NewDecFromInt(amount)
	feeRate := types.NewDec(int64(imm.FeeRate)).Quo(types.NewDec(10000))
	fees := amountDec.Mul(feeRate)
	amountMinusFees := amountDec.Sub(fees)
	return amountMinusFees
}

// Worth Function V=M)
func (imm *InterchainMarketMaker) Invariant() float64 {
	v := 1.0
	totalBalance := types.NewDec(0)
	for _, pool := range imm.Pool.Assets {
		totalBalance = totalBalance.Add(types.NewDecFromBigInt(pool.Balance.Amount.BigInt()))
	}
	for _, pool := range imm.Pool.Assets {
		w := float64(pool.Weight) / 100.0
		balance := types.NewDecFromBigInt(pool.Balance.Amount.BigInt()).Quo(totalBalance) /// totalBalance
		v *= math.Pow(balance.MustFloat64(), w)
	}
	return v
}

func (imm *InterchainMarketMaker) InvariantWithInput(tokenIn types.Coin) float64 {
	v := 1.0
	totalBalance := types.NewDec(0)
	for _, pool := range imm.Pool.Assets {
		totalBalance = totalBalance.Add(types.NewDecFromBigInt(pool.Balance.Amount.BigInt()))
		if pool.Balance.Denom == tokenIn.Denom {
			totalBalance.Add(types.NewDecFromBigInt(tokenIn.Amount.BigInt()))
		}
	}
	for _, pool := range imm.Pool.Assets {
		w := float64(pool.Weight) / 100.0
		var balance types.Dec
		if tokenIn.Denom != pool.Balance.Denom {
			balance = types.NewDecFromBigInt(pool.Balance.Amount.BigInt()).Quo(totalBalance) /// totalBalance
		} else {
			balance = types.NewDecFromBigInt(pool.Balance.Amount.Add(tokenIn.Amount).BigInt()).Quo(totalBalance) /// totalBalance
		}
		v *= math.Pow(balance.MustFloat64(), w)
	}
	return v
}

func (imm *InterchainMarketMaker) LpPrice() float64 {
	lpPrice := imm.Invariant() / float64(imm.Pool.Supply.Amount.Int64())
	return lpPrice
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
	worth := amm.Invariant()
	lpPrice := amm.LpPrice() // worth * 1e10 / float64(amm.Pool.Supply.Amount.Int64())

	outA, _ := amm.Withdraw(types.NewCoin(amm.Pool.PoolId, userALp), denomA)
	outB, _ := amm.Withdraw(types.NewCoin(amm.Pool.PoolId, userBLp), denomB)

	fmt.Println("----*********----")
	fmt.Println("UserALPAmount:", userALp)
	fmt.Println("UserBLPAmount:", userBLp)
	fmt.Println("WithdrawUSDT:", outA)
	fmt.Println("WithdrawETH:", outB)
	fmt.Println("Worth:", worth)
	fmt.Println("LPprice:", lpPrice)
	// fmt.Println("--------------------------------------------\n")
}
