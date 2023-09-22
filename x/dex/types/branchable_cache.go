package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type BranchableCache struct {
	Parent     *BranchableCache
	Ctx        sdk.Context
	writeCache func()
}

func (bc *BranchableCache) Branch() *BranchableCache {
	cacheCtx, writeCache := bc.Ctx.CacheContext()
	return &BranchableCache{
		Parent:     bc,
		Ctx:        cacheCtx,
		writeCache: writeCache,
	}
}

func (bc *BranchableCache) WriteCache() {
	bc.writeCache()
	if bc.Parent != nil {
		bc.Parent.WriteCache()
	}
}

func NewBranchableCache(ctx sdk.Context) *BranchableCache {
	cacheCtx, writeCache := ctx.CacheContext()
	return &BranchableCache{
		Parent:     nil,
		Ctx:        cacheCtx,
		writeCache: writeCache,
	}
}
