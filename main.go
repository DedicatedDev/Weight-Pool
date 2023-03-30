package main

import (
	"fmt"

	types "github.com/cosmos/cosmos-sdk/types"
	market "github.com/dedicatedDev/WeightPool/market"
)

const initialX = 2_000_000 // USDT
const initialY = 1000      // ETH
const denomA = "USDT"
const denomB = "ETH"
const fee = 300

func main() {

	userALpAmount := types.NewInt((initialX + initialY) / 2)
	userBLpAmount := types.NewInt((initialX + initialY) / 2)

	// make pool
	pool := market.NewInterchainLiquidityPool(
		"creator",
		[]*types.Coin{
			{Denom: denomA, Amount: types.NewInt(initialX)},
			{Denom: denomB, Amount: types.NewInt(initialY)}},
		[]uint32{6, 6},
		"50:50",
		"test",
		"test",
	)
	pool.AddPoolSupply(types.NewCoin(
		pool.PoolId,
		types.NewInt(initialX+initialY),
	))

	// make market maker
	amm := market.NewInterchainMarketMaker(
		pool,
		fee,
	)

	// check current price.
	amm.LogPrice("Create Pool", denomA, denomB)

	// going test process
	// ================================
	depositCoin := types.NewCoin(denomA, types.NewInt(100))
	outToken, _ := amm.DepositSingleAsset(depositCoin)
	userALpAmount = userALpAmount.Add(outToken.Amount)
	amm.LogPrice("Step0: deposit Asset A (100)", denomA, denomB)
	fmt.Println("LPAmount", userALpAmount)
	out, _ := amm.Withdraw(*outToken, denomA)
	fmt.Println("USDT", out)

	// ================================
	depositCoin = types.NewCoin(denomA, types.NewInt(initialX))
	outToken, _ = amm.DepositSingleAsset(depositCoin)
	userALpAmount = userALpAmount.Add(outToken.Amount)
	amm.LogPrice("Step1: deposit Asset A (initialX)", denomA, denomB)

	// ================================
	depositCoin = types.NewCoin(
		denomB,
		types.NewInt(initialY),
	)

	outToken, _ = amm.DepositSingleAsset(depositCoin)
	userBLpAmount = userBLpAmount.Add(outToken.Amount)
	amm.LogPrice("Step2: deposit Asset B (initialY)", denomA, denomB)

	// ================================
	tokenIn := types.NewCoin(denomA, types.NewInt(10_0000))
	tokenOut, _ := amm.LeftSwap(tokenIn, denomB)

	pool.AddAsset(tokenIn)
	pool.SubAsset(*tokenOut)
	amm.LogPrice("step3: swap Asset A(10_0000) to Asset B", denomA, denomB)

	// ================================
	tokenIn = types.NewCoin(denomA, types.NewInt(100_0000))
	tokenOut, _ = amm.LeftSwap(tokenIn, denomB)

	pool.AddAsset(tokenIn)
	pool.SubAsset(*tokenOut)
	amm.LogPrice("step4: swap Asset A(100_0000) to Asset B", denomA, denomB)

	// ================================
	depositCoin = types.NewCoin(denomA, types.NewInt(100_000))
	outToken, _ = amm.DepositSingleAsset(depositCoin)
	userALpAmount = userALpAmount.Add(outToken.Amount)
	amm.LogPrice("step5: deposit Asset A (100_000)", denomA, denomB)

	// ================================
	outToken, _ = amm.DepositSingleAsset(depositCoin)
	userALpAmount = userALpAmount.Add(outToken.Amount)
	amm.LogPrice("step6: withdraw Asset A (50%)", denomA, denomB)

	userAPoolCoin := types.NewCoin(
		pool.PoolId,
		userALpAmount,
	)

	userBPoolCoin := types.NewCoin(
		pool.PoolId,
		userBLpAmount,
	)
	fmt.Println(userAPoolCoin)
	fmt.Println(userBPoolCoin)
	withdrawATokenAmount, _ := amm.Withdraw(userAPoolCoin, denomA)
	withdrawBTokenAmount, _ := amm.Withdraw(userBPoolCoin, denomB)
	fmt.Println(withdrawATokenAmount)
	fmt.Println(withdrawBTokenAmount)
}
