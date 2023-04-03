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
	creatorALpAmount := initialLPSupply
	creatorBLpAmount := initialLPSupply

	userALpAmount := types.NewDec((0))
	userBLpAmount := types.NewDec((0))

	// make pool
	pool := market.NewInterchainLiquidityPool(
		"creator",
		[]*types.DecCoin{
			{Denom: denomA, Amount: types.NewDec(initialX)},
			{Denom: denomB, Amount: types.NewDec(initialY)}},
		[]uint32{6, 6},
		"20:80",
		"test",
		"test",
	)
	pool.AddPoolSupply(types.NewDecCoin(
		pool.PoolId,
		initialLPSupply,
	))

	// make market maker
	amm := market.NewInterchainMarketMaker(
		pool,
		fee,
	)

	tvl := amm.TVL()
	lpPrice := tvl * 1e10 / pool.Supply.Amount.MustFloat64()
	pool.PoolTokenPrice = lpPrice

	// check current price.
	amm.LogPrice("Create Pool", denomA, denomB, creatorALpAmount, creatorBLpAmount)

	// going test process
	// ================================
	depositCoin := types.NewDecCoin(denomA, types.NewInt(initialX))
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
	amm.LogPrice("Step0: deposit Asset A (initialX)", denomA, denomB, userALpAmount.RoundInt(), creatorBLpAmount)

	//================================
	depositCoin = types.NewDecCoin(denomB, types.NewInt(initialY))
	//depositCoin, _ = amm.CheckMaxDepositAmount(denomB, 0.1)

	fmt.Println("DepositAmount:", depositCoin)
	outToken, err = amm.DepositSingleAsset(depositCoin)
	if err != nil {
		fmt.Println(err)
		return
	}
	userBLpAmount = userBLpAmount.Add(outToken.Amount)
	amm.LogPrice("Step2: deposit Asset B (initialY) By User B", denomA, denomB, userALpAmount.RoundInt(), userBLpAmount.RoundInt())

	// ================================
	tokenIn := types.NewDecCoin(denomB, types.NewInt(100))
	tokenOut, _ := amm.LeftSwap(tokenIn, denomA)
	pool.AddAsset(tokenIn)
	pool.SubAsset(*tokenOut)
	amm.LogPrice("step3: swap Asset A(10_0000) to Asset B", denomA, denomB, userALpAmount.RoundInt(), userBLpAmount.RoundInt())

	// // ================================
	// depositCoin, _ = amm.CheckMaxDepositAmount(denomB, 0.1)
	// fmt.Println("Current Allow Single Deposit Amount:", depositCoin)

	// outToken, _ = amm.DepositSingleAsset(*depositCoin)
	// userBLpAmount = userBLpAmount.Add(outToken.Amount)
	// amm.LogPrice("step4: make pool as a balance", denomA, denomB, userALpAmount.RoundInt(), userBLpAmount.RoundInt())

	// // ================================
	// tokenIn = types.NewDecCoin(denomA, types.NewInt(100_0000))
	// tokenOut, _ = amm.LeftSwap(tokenIn, denomB)
	// pool.AddAsset(tokenIn)
	// pool.SubAsset(*tokenOut)
	// amm.LogPrice("step4: swap Asset A(100_0000) to Asset B", denomA, denomB, userALpAmount.RoundInt(), userBLpAmount.RoundInt())

	// ================================
	tokenIn = types.NewDecCoin(denomB, types.NewInt(100_0000))
	tokenOut, _ = amm.LeftSwap(tokenIn, denomB)
	pool.AddAsset(tokenIn)
	pool.SubAsset(*tokenOut)
	amm.LogPrice("step4: swap Asset B(100_0000) to Asset B", denomA, denomB, userALpAmount.RoundInt(), userBLpAmount.RoundInt())

	// ================================
	depositCoin = types.NewDecCoin(denomA, types.NewInt(100_000))
	outToken, _ = amm.DepositSingleAsset(depositCoin)
	userALpAmount = userALpAmount.Add(outToken.Amount)
	amm.LogPrice("step5: deposit Asset A (100_000)", denomA, denomB, userALpAmount.RoundInt(), userBLpAmount.RoundInt())

	// ================================
	depositCoin = types.NewDecCoin(denomA, types.NewInt(100_000))
	outToken, _ = amm.DepositSingleAsset(depositCoin)
	userALpAmount = userALpAmount.Add(outToken.Amount)
	amm.LogPrice("step6: deposit Asset A (100_000)", denomA, denomB, userALpAmount.RoundInt(), userBLpAmount.RoundInt())

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
