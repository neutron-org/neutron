package types

import "fmt"

// ValidateHookType ensures all hooks in the slice exist and are not unspecified
func ValidateHookType(hook HookType) error {
	_, ok := HookType_name[int32(hook)]
	if !ok {
		return fmt.Errorf("non-existing hook=%d", int32(hook))
	}

	if hook == HOOK_TYPE_UNSPECIFIED {
		return fmt.Errorf("unspecified hooks are not allowed")
	}

	return nil
}
