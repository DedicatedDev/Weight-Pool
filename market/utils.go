package market

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetPoolId(denoms []string) string {
	//generate poolId
	sort.Strings(denoms)
	poolIdHash := sha256.New()
	poolIdHash.Write([]byte(strings.Join(denoms, "")))
	poolId := "pool" + fmt.Sprintf("%v", hex.EncodeToString(poolIdHash.Sum(nil)))
	//poolId := "pool" + fmt.Sprintf("%v", hex.EncodeToString(poolIdHash.Sum(nil)))
	return poolId
}

func GetPoolIdWithTokens(tokens []*sdk.DecCoin) string {

	denoms := []string{}
	for _, token := range tokens {
		denoms = append(denoms, token.Denom)
	}
	return GetPoolId(denoms)
}
