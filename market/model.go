package market

import (
	types "github.com/cosmos/cosmos-sdk/types"
)

type PoolSide int32

const (
	PoolSide_NATIVE PoolSide = 0
	PoolSide_REMOTE PoolSide = 1
)

type PoolStatus int32

const (
	PoolStatus_POOL_STATUS_INITIAL PoolStatus = 0
	PoolStatus_POOL_STATUS_READY   PoolStatus = 1
)

type PoolAsset struct {
	Side    PoolSide    `protobuf:"varint,1,opt,name=side,proto3,enum=ibc.applications.interchain_swap.v1.PoolSide" json:"side,omitempty"`
	Balance *types.Coin `protobuf:"bytes,2,opt,name=balance,proto3" json:"balance,omitempty"`
	Weight  uint32      `protobuf:"varint,3,opt,name=weight,proto3" json:"weight,omitempty"`
	Decimal uint32      `protobuf:"varint,4,opt,name=decimal,proto3" json:"decimal,omitempty"`
}

type InterchainLiquidityPool struct {
	PoolId                string       `protobuf:"bytes,1,opt,name=poolId,proto3" json:"poolId,omitempty"`
	Creator               string       `protobuf:"bytes,2,opt,name=creator,proto3" json:"creator,omitempty"`
	Assets                []*PoolAsset `protobuf:"bytes,3,rep,name=assets,proto3" json:"assets,omitempty"`
	Supply                *types.Coin  `protobuf:"bytes,4,opt,name=supply,proto3" json:"supply,omitempty"`
	Status                PoolStatus   `protobuf:"varint,5,opt,name=status,proto3,enum=ibc.applications.interchain_swap.v1.PoolStatus" json:"status,omitempty"`
	EncounterPartyPort    string       `protobuf:"bytes,6,opt,name=encounterPartyPort,proto3" json:"encounterPartyPort,omitempty"`
	EncounterPartyChannel string       `protobuf:"bytes,7,opt,name=encounterPartyChannel,proto3" json:"encounterPartyChannel,omitempty"`
}

type InterchainMarketMaker struct {
	PoolId  string                   `protobuf:"bytes,1,opt,name=poolId,proto3" json:"poolId,omitempty"`
	Pool    *InterchainLiquidityPool `protobuf:"bytes,2,opt,name=pool,proto3" json:"pool,omitempty"`
	FeeRate uint32                   `protobuf:"varint,3,opt,name=feeRate,proto3" json:"feeRate,omitempty"`
}
