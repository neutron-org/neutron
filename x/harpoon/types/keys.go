package types

const (
	// ModuleName defines the module name
	ModuleName = "harpoon"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_harpoon"

    
)

var (
	ParamsKey = []byte("p_harpoon")
)



func KeyPrefix(p string) []byte {
    return []byte(p)
}
