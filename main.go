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

	initialLPSupply := types.NewInt(int64((initialX + initialY)))
	creatorALpAmount := initialLPSupply.Mul(types.NewInt(20)).Quo(types.NewInt(100))
	creatorBLpAmount := initialLPSupply.Mul(types.NewInt(80)).Quo(types.NewInt(100))
	fmt.Println(creatorALpAmount, creatorBLpAmount)

	userALpAmount := types.NewInt((0))
	userBLpAmount := types.NewInt((0))

	// make pool
	pool := market.NewInterchainLiquidityPool(
		"creator",
		[]*types.Coin{
			{Denom: denomA, Amount: types.NewInt(initialX)},
			{Denom: denomB, Amount: types.NewInt(initialY)}},
		[]uint32{6, 6},
		"20:80",
		"test",
		"test",
	)
	pool.AddPoolSupply(types.NewCoin(
		pool.PoolId,
		initialLPSupply,
	))

	// make market maker
	amm := market.NewInterchainMarketMaker(
		pool,
		fee,
	)

	//tvl := amm.Invariant()
	//lpPrice := tvl / float64(pool.Supply.Amount.Int64())
	pool.PoolPrice = amm.LpPrice()

	// check current price.
	amm.LogPrice("Create Pool", denomA, denomB, creatorALpAmount, creatorBLpAmount)

	// going test process
	// ================================
	depositCoin := types.NewCoin(denomA, types.NewInt(initialX*0.01))
	//depositCoin, _ := amm.CheckMaxDepositAmount(denomA, 0.1)
	fmt.Println("DepositAmount:", depositCoin)
	//fmt.Printf(err.Error())
	//fmt.Print(coin)
	outToken, err := amm.DepositSingleAsset(depositCoin)
	if err != nil {
		fmt.Println(err)
		return
	}
	userALpAmount = userALpAmount.Add(outToken.Amount)
	amm.LogPrice("Step0: deposit Asset A (initialX)", denomA, denomB, userALpAmount, creatorBLpAmount)

	//================================
	depositCoin = types.NewCoin(denomB, types.NewInt(initialY))
	//depositCoin, _ = amm.CheckMaxDepositAmount(denomB, 0.1)

	fmt.Println("DepositAmount:", depositCoin)
	outToken, err = amm.DepositSingleAsset(depositCoin)
	if err != nil {
		fmt.Println(err)
		return
	}
	userBLpAmount = userBLpAmount.Add(outToken.Amount)
	amm.LogPrice("Step2: deposit Asset B (initialY) By User B", denomA, denomB, userALpAmount, userBLpAmount)

	// ================================
	tokenIn := types.NewCoin(denomB, types.NewInt(100))
	tokenOut, _ := amm.LeftSwap(tokenIn, denomA)
	pool.AddAsset(tokenIn)
	pool.SubAsset(*tokenOut)
	amm.LogPrice("step3: swap Asset A(10_0000) to Asset B", denomA, denomB, userALpAmount, userBLpAmount)

	// // ================================
	// depositCoin, _ = amm.CheckMaxDepositAmount(denomB, 0.1)
	// fmt.Println("Current Allow Single Deposit Amount:", depositCoin)

	// outToken, _ = amm.DepositSingleAsset(*depositCoin)
	// userBLpAmount = userBLpAmount.Add(outToken.Amount)
	// amm.LogPrice("step4: make pool as a balance", denomA, denomB, userALpAmount.RoundInt(), userBLpAmount.RoundInt())

	// // ================================
	// tokenIn = types.NewCoin(denomA, types.NewInt(100_0000))
	// tokenOut, _ = amm.LeftSwap(tokenIn, denomB)
	// pool.AddAsset(tokenIn)
	// pool.SubAsset(*tokenOut)
	// amm.LogPrice("step4: swap Asset A(100_0000) to Asset B", denomA, denomB, userALpAmount, userBLpAmount.RoundInt())

	// // ================================
	// tokenIn = types.NewCoin(denomB, types.NewInt(100_0000))
	// tokenOut, _ = amm.LeftSwap(tokenIn, denomB)
	// pool.AddAsset(tokenIn)
	// pool.SubAsset(*tokenOut)

	// amm.LogPrice("step4: swap Asset B(100_0000) to Asset B", denomA, denomB, userALpAmount, userBLpAmount)

	// // ================================
	// depositCoin = types.NewCoin(denomA, types.NewInt(100_000))
	// outToken, err = amm.DepositSingleAsset(depositCoin)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	userALpAmount = userALpAmount.Add(outToken.Amount)
	// 	amm.LogPrice("step5: deposit Asset A (100_000)", denomA, denomB, userALpAmount, userBLpAmount)
	// }

	// // ================================
	// depositCoin = types.NewCoin(denomA, types.NewInt(100_000))
	// outToken, _ = amm.DepositSingleAsset(depositCoin)
	// userALpAmount = userALpAmount.Add(outToken.Amount)
	// amm.LogPrice("step6: deposit Asset A (100_000)", denomA, denomB, userALpAmount, userBLpAmount)

	// // ================================
	// outToken, _ = amm.DepositSingleAsset(depositCoin)
	// userALpAmount = userALpAmount.Add(outToken.Amount)
	// amm.LogPrice("step7: withdraw Asset", denomA, denomB, userALpAmount.RoundInt(), userBLpAmount.RoundInt())

	// userAPoolCoin := types.NewDecCoin(
	// 	pool.PoolId,
	// 	userALpAmount.RoundInt(),
	// )

	// userBPoolCoin := types.NewCoin(
	// 	pool.PoolId,
	// 	userBLpAmount,
	// )
	// fmt.Println(userAPoolCoin)
	// fmt.Println(userBPoolCoin)
	// withdrawATokenAmount, _ := amm.Withdraw(userAPoolCoin, denomA)
	// withdrawBTokenAmount, _ := amm.Withdraw(userBPoolCoin, denomB)
	// fmt.Println(withdrawATokenAmount)
	// fmt.Println(withdrawBTokenAmount)
}
